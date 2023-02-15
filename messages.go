package resmix

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/jonathanMweiss/resmix/config"
	"github.com/jonathanMweiss/resmix/internal/crypto/tibe"
	"golang.org/x/crypto/sha3"
	"math"
	"os"
	"runtime"
)

const defaultMessageStoreageLocation = "message.storage"
const messageSize = 258

type MessageGenerator struct {
	*config.SystemConfig
	messageStoreageLocation string
	mpks                    map[string]*tibe.MasterPublicKey
}

type Onion []byte

func (c Onion) ExtractMixConfig(topology *config.Topology) *config.LogicalMix {
	n := len(c)
	layer := binary.BigEndian.Uint32(c[n-8 : n-4])
	posInLayer := binary.BigEndian.Uint32(c[n-4:])

	return topology.Layers[layer].LogicalMixes[posInLayer]
}

func (c Onion) ExtractCipher() *tibe.Cipher {
	cphr := &tibe.Cipher{}
	if err := cphr.SetBytes(c[:len(c)-8]); err != nil {
		panic("invalid onion")
	}

	return cphr
}

func GroupOnionsByMixName(onions []Onion, topology *config.Topology) map[string][]Onion {
	numServers := len(topology.Layers[0].LogicalMixes)

	m := map[string][]Onion{}

	for _, o := range onions {
		mix := o.ExtractMixConfig(topology)

		if _, ok := m[mix.Name]; !ok {
			m[mix.Name] = make([]Onion, 0, len(onions)/numServers)
		}

		m[mix.Name] = append(m[mix.Name], o)
	}

	return m
}

func GroupOnionsByHostname(onions []Onion, topology *config.Topology) map[string][]Onion {
	numServers := len(topology.Layers[0].LogicalMixes)

	m := map[string][]Onion{}

	for _, o := range onions {
		mix := o.ExtractMixConfig(topology)
		address := mix.Hostname

		if _, ok := m[address]; !ok {
			m[address] = make([]Onion, 0, len(onions)/numServers)
		}

		m[address] = append(m[address], o)
	}

	return m
}

func (m MessageGenerator) onionWrap(
	round int, layer, posInLayer uint32, msg []byte, serverMasterPublicKey *tibe.MasterPublicKey) (Onion, error) {
	sname := m.SystemConfig.Topology.Layers[layer].LogicalMixes[posInLayer].Hostname

	id := computeId(sname, round)

	cipher, err := serverMasterPublicKey.Encrypt(id[:], msg)
	if err != nil {
		return Onion{}, err
	}

	bf := bytes.NewBuffer(make([]byte, 0, cipher.Size()+4*2))
	cipher.ToBuffer(bf)

	numBuf := [4]byte{}

	binary.BigEndian.PutUint32(numBuf[:], layer)
	bf.Write(numBuf[:])

	binary.BigEndian.PutUint32(numBuf[:], posInLayer)
	bf.Write(numBuf[:])

	return bf.Bytes(), nil
}

func NewMessageGenerator(sysConfigs *config.SystemConfig) *MessageGenerator {
	m := &MessageGenerator{
		SystemConfig:            sysConfigs,
		messageStoreageLocation: defaultMessageStoreageLocation,
		mpks:                    make(map[string]*tibe.MasterPublicKey),
	}

	for s, mpkBytes := range sysConfigs.ServerConfigs[0].IBEConfigs.AddressOfNodeToMasterPublicKeys {
		mpk := &tibe.MasterPublicKey{}
		if err := mpk.SetBytes(mpkBytes); err != nil {
			panic(err)
		}

		m.mpks[s] = mpk
	}

	return m
}

// LoadOrCreateMessages returns numclients * sqrt(2*numChains) onions
func (m *MessageGenerator) LoadOrCreateMessages(numClients, round int) ([]Onion, error) {
	fname := fmt.Sprintf("%v.%v", numClients, m.messageStoreageLocation)
	_, err := os.Stat(fname)
	if err == nil {
		return m.loadMessages(fname)
	}

	onions := m.generateOnions(numClients, round)

	bts, err := json.Marshal(onions)
	if err != nil {
		return nil, err
	}

	if err := os.WriteFile(fname, bts, 0777); err != nil {
		return nil, err
	}

	return onions, nil
}

func (m *MessageGenerator) loadMessages(fname string) ([]Onion, error) {
	content, err := os.ReadFile(fname)
	if err != nil {
		return nil, err
	}

	onions := make([]Onion, 0)
	if err := json.Unmarshal(content, &onions); err != nil {
		return nil, err
	}

	return onions, nil
}

func (m *MessageGenerator) generateOnions(numClients int, round int) []Onion {
	onions := make([]Onion, 0, numClients*int(math.Pow(float64(len(m.SystemConfig.Topology.Layers)), 2)))

	N := runtime.NumCPU()
	results := make(chan []Onion, 2*N)

	partitionSize := numClients / N

	current := 0

	for i := 0; i < N; i++ {
		var numClientsPerThread int
		if i+1 == N {
			numClientsPerThread = numClients - current
		} else {
			numClientsPerThread = partitionSize
		}

		current += partitionSize

		go func() {
			for i := 0; i < numClientsPerThread; i++ {
				randomMsg := make([]byte, messageSize)
				_, _ = rand.Read(randomMsg)

				results <- m.generateOnion(randomMsg, round)
			}
		}()

	}

	for i := 0; i < numClients; i++ {
		onions = append(onions, <-results...)
	}

	return onions
}

func (m *MessageGenerator) generateOnion(msg []byte, round int) []Onion {

	// choose random mixes in each layer. // do so according to the random message. this is a POC.
	numberOfChains := len(m.SystemConfig.Topology.Layers[0].LogicalMixes)

	chosenChains := calculateChain(numberOfChains, msg)

	onions := make([]Onion, 0, len(chosenChains))
	for _, chain := range chosenChains {
		//create onion
		layers := m.SystemConfig.Topology.Layers

		onion := msg
		var err error

		for i := len(layers) - 1; i >= 0; i-- {
			currentMix := layers[i].LogicalMixes[chain]
			onion, err = m.onionWrap(round, uint32(i), uint32(chain), onion, m.mpks[currentMix.Hostname])
			if err != nil {
				panic(err)
			}

		}

		// store onion
		onions = append(onions, onion)
	}

	return onions
}

func calculateChain(numberOfChains int, randomMsg []byte) []int {
	numberOfRepetitions := int(math.Ceil(math.Sqrt(2 * float64(numberOfChains))))

	chosenChains := make([]int, 0, numberOfRepetitions)

	current := sha3.Sum256(randomMsg)
	row := binary.BigEndian.Uint64(current[:]) % uint64(numberOfChains)
	chosenChains = append(chosenChains, int(row))

	for i := 1; i < numberOfRepetitions; i++ {
		current = sha3.Sum256(current[:])
		row := binary.BigEndian.Uint64(current[:]) % uint64(numberOfChains)
		chosenChains = append(chosenChains, int(row))
	}

	return chosenChains
}
