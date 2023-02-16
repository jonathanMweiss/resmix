package resmix

import (
	"crypto/rand"
	"github.com/algorand/go-deadlock"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"runtime"
	"sync"

	"github.com/jonathanMweiss/resmix/config"
	"github.com/jonathanMweiss/resmix/internal/crypto/tibe"
)

// MixHandler ensures sequencial access. Any method called will block until the previous method has finished.
type MixHandler interface {
	// UpdateMixes states a failure and adds information regarding the new topology, keys etc.
	UpdateMixes(recoveryScheme)

	// AddMessages adds messages to a LogicalMix.
	AddMessages(messages []*Messages)

	// GetOutputsChan returns processed messages.
	GetOutputsChan() <-chan mixoutput

	GetMixOutputs(mxName mixName) ([]Onion, error)

	Close()
}

type recoveryScheme struct {
	newTopology *config.Topology

	newResponsibility map[mixName]hostname
	keys              map[mixName]tibe.Decrypter
}

type mixoutput struct {
	onions        []Onion
	logicalSender []byte
}

// Mixers implements MixHandler
type Mixers struct {
	states    map[mixName]*mixState
	tasksChan chan *decryptionTask
	log       *logrus.Entry
	output    chan mixoutput

	mu deadlock.RWMutex
}

type decryptionTask struct {
	wg *sync.WaitGroup

	tibe.Decrypter
	input Onion

	// output onion on the
	Idx    int
	Result []Onion
}

type physyicalMixName struct {
	Hostname hostname
	MixName  mixName
}

type mixState struct {
	name string

	decryptionKey tibe.Decrypter

	predecessors  map[physyicalMixName]int // usually one, unless the host of that mix failed.
	successors    []mixName                // usually one, unless the host of that mix failed.
	totalReceived map[physyicalMixName]int // total number of messages received from each predecessor.

	totalWorkload int
	outputs       []Onion // should be of size totalWorkload

	wg          sync.WaitGroup // used to wait for all messages to be processed by the threadpool.
	shuffleWait sync.WaitGroup // used to wait for the shuffler to finish.
	shuffler    *Shuffler

	sync.Once

	isLastLayer bool
}

func NewMixers(topo *config.Topology, mixesConfigs []*config.LogicalMix, decrypter tibe.Decrypter, workloadMap map[mixName]int) MixHandler {
	m := &Mixers{
		log:       logrus.WithField("component", "mixer"),
		states:    make(map[mixName]*mixState),
		tasksChan: make(chan *decryptionTask, 1000), // TODO: discuss size with Yossi. i assume we'll want the buffer to fit the max workload of a single mix.
		output:    make(chan mixoutput, len(mixesConfigs)*2),
	}

	for i := 0; i < runtime.NumCPU(); i++ {
		// Decryption workers
		go func() {
			for task := range m.tasksChan {
				cipher := task.input.ExtractCipher()
				outOnion, err := task.Decrypter.Decrypt(*cipher)
				if err != nil {
					m.log.Errorln("failed to decrypt message", err)
				}

				// either nil/ or the decrypted message.
				task.Result[task.Idx] = outOnion
				task.wg.Done()
			}
		}()
	}

	for _, mix := range mixesConfigs {
		m.states[mixName(mix.Name)] = newMixState(topo, mix, decrypter, workloadMap)
	}

	return m
}

func (m *Mixers) Close() {
	close(m.tasksChan)
}

func newMixState(topo *config.Topology, mix *config.LogicalMix, decrypter tibe.Decrypter, workloadMap map[mixName]int) *mixState {
	mx := &mixState{
		name: mix.Name,

		decryptionKey: decrypter,

		predecessors:  make(map[physyicalMixName]int),
		successors:    make([]mixName, len(mix.Successors)),
		totalReceived: make(map[physyicalMixName]int),
		totalWorkload: workloadMap[mixName(mix.Name)],

		outputs: make([]Onion, workloadMap[mixName(mix.Name)]),

		wg:          sync.WaitGroup{},
		shuffleWait: sync.WaitGroup{},
		shuffler:    NewShuffler(rand.Reader),
		Once:        sync.Once{},

		isLastLayer: int(mix.Layer) == len(topo.Layers)-1,
	}

	mx.shuffleWait.Add(1)       // ensuring we wait for shuffle to finish.
	mx.wg.Add(mx.totalWorkload) // ensuring we wait for all messages to be processed.

	for i, successor := range mix.Successors {
		mx.successors[i] = mixName(successor)
	}

	mx.SetPredecessors(topo, mix, workloadMap)

	return mx
}

func (m *mixState) SetPredecessors(topo *config.Topology, mix *config.LogicalMix, workloadMap map[mixName]int) {
	for _, predecessor := range mix.Predecessors {
		host := hostname(config.GenesisName)
		workload := m.totalWorkload
		if predecessor != config.GenesisName {
			host = hostname(topo.Mixes[predecessor].Hostname)
			workload = workloadMap[mixName(predecessor)]
		}

		pname := physyicalMixName{
			Hostname: host,
			MixName:  mixName(predecessor),
		}

		m.predecessors[pname] = workload
	}

	for name := range m.predecessors {
		m.totalReceived[name] = 0
	}

}

func (m *mixState) GetOutputs() []Onion {
	m.wg.Wait()

	m.shuffleWait.Wait()

	return m.outputs
}

func (m *Mixers) AddMessages(messages []*Messages) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, message := range messages {
		mix := m.states[mixName(message.LogicalReceiver)]

		decryptionTasks := mix.HandleNewInputs(message)

		for _, task := range decryptionTasks {
			m.tasksChan <- task
		}

		if !mix.isDone() {
			continue
		}

		go m.shuffleAndSend(mix)
	}
}

// HandleNewInputs updates state, and creates decryption tasks if the workload is complete.
func (m *mixState) HandleNewInputs(message *Messages) []*decryptionTask {
	tsks := make([]*decryptionTask, len(message.Messages))

	// todo: consider some verification on input. like - not processing the same messages twice...
	for i, bytes := range message.Messages {
		tsks[i] = &decryptionTask{
			wg:        &m.wg,
			Decrypter: m.decryptionKey,
			input:     bytes,
			Idx:       m.totalWorkload - 1,
			Result:    m.outputs,
		}

		m.totalWorkload -= 1
	}

	m.totalReceived[physyicalMixName{
		Hostname: hostname(message.PhysicalSender),
		MixName:  mixName(message.LogicalSender),
	}] += len(message.Messages)

	return tsks
}

func (m *mixState) isDone() bool {
	for k, v := range m.predecessors {
		if v != m.totalReceived[k] {
			return false
		}
	}

	return true
}

func (m *mixState) signalDone() {
	m.shuffleWait.Done()
}

func (m *Mixers) shuffleAndSend(mix *mixState) {
	mix.wg.Wait()

	mix.Once.Do(
		func() {
			mix.shuffler.ShuffleOnions(mix.outputs)

			mix.signalDone()
		},
	)

	// do not send the output to the next layer if this is the last layer.
	if mix.isLastLayer {
		return
	}

	m.output <- mixoutput{
		mix.GetOutputs(),
		[]byte(mix.name),
	}
}

func (m *Mixers) GetMixOutputs(mxName mixName) ([]Onion, error) {
	m.mu.RLock()
	mx, ok := m.states[mixName(mxName)]
	m.mu.RUnlock()

	if !ok {
		return nil, status.Error(codes.NotFound, "mix not found")
	}

	return mx.GetOutputs(), nil
}

func (m *Mixers) UpdateMixes(scheme recoveryScheme) {
	m.mu.Lock()
	defer m.mu.Unlock()

	//TODO implement me
	panic("implement me")
}

func (m *Mixers) GetOutputsChan() <-chan mixoutput {
	return m.output
}
