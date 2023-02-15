package resmix

import (
	"crypto/rand"
	"github.com/sirupsen/logrus"
	"runtime"
	"sync"

	"github.com/jonathanMweiss/resmix/config"
	"github.com/jonathanMweiss/resmix/internal/crypto/tibe"
)

type Mixers struct {
	states    map[mixName]*mixState
	tasksChan chan *decryptionTask
	log       *logrus.Entry
	output    chan mixoutput
}

func (m *Mixers) Close() {
	close(m.tasksChan)
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
	shuffler    *Shuffler
	isLastLayer bool
}

// createDecryptionTasks updates state, and creates decryption tasks if the workload is complete.
func (m *mixState) createDecryptionTasks(message *Messages) []*decryptionTask {
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
	// TODO
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

func NewMixers(topo *config.Topology, mixesConfigs []*config.LogicalMix, decrypter tibe.Decrypter, workloadMap map[mixName]int) MixHandler {
	m := &Mixers{
		log:       logrus.WithField("component", "mixer"),
		states:    make(map[mixName]*mixState),
		tasksChan: make(chan *decryptionTask, 1000),
		output:    make(chan mixoutput, len(mixesConfigs)*2),
	}

	for i := 0; i < runtime.NumCPU(); i++ {
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

func newMixState(topo *config.Topology, mix *config.LogicalMix, decrypter tibe.Decrypter, workloadMap map[mixName]int) *mixState {
	mx := &mixState{
		name:          mix.Name,
		decryptionKey: decrypter,
		predecessors:  make(map[physyicalMixName]int),
		successors:    make([]mixName, len(mix.Successors)),
		totalReceived: make(map[physyicalMixName]int),
		totalWorkload: workloadMap[mixName(mix.Name)],
		outputs:       make([]Onion, workloadMap[mixName(mix.Name)]),
		wg:            sync.WaitGroup{},
		shuffler:      NewShuffler(rand.Reader),

		isLastLayer: int(mix.Layer) == len(topo.Layers)-1,
	}

	for i, successor := range mix.Successors {
		mx.successors[i] = mixName(successor)
	}

	mx.SetPredecessors(topo, mix, workloadMap)

	mx.wg.Add(mx.totalWorkload)

	return mx
}

type decryptionTask struct {
	wg *sync.WaitGroup

	tibe.Decrypter
	input Onion

	// output onion on the
	Idx    int
	Result []Onion
}

func (m *Mixers) AddMessages(messages []*Messages) {
	for _, message := range messages {
		mix := m.states[mixName(message.LogicalReceiver)]

		decryptionTasks := mix.createDecryptionTasks(message)

		for _, task := range decryptionTasks {
			m.tasksChan <- task
		}

		if !mix.isDone() {
			continue
		}

		go func() {
			mix.wg.Wait()

			mix.shuffler.ShuffleOnions(mix.outputs)

			mix.signalDone()

			// do not send the output to the next layer if this is the last layer.
			if mix.isLastLayer {
				return
			}

			m.output <- mixoutput{
				mix.outputs,
				[]byte(mix.name),
			}
		}()
	}
}

func (m *Mixers) UpdateMixes(scheme recoveryScheme) {
	//TODO implement me
	panic("implement me")
}

// Someone must wait on this?
func (m *Mixers) GetOutputsChan() <-chan mixoutput {
	return m.output
}
