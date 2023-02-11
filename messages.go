package resmix

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"github.com/jonathanMweiss/resmix/config"
	"github.com/jonathanMweiss/resmix/internal/crypto/tibe"
	"golang.org/x/crypto/sha3"
	"math"
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

func (c Onion) ExtractMixName(topology *config.Topology) string {
	n := len(c)
	layer := binary.BigEndian.Uint32(c[n-8 : n-4])
	posInLayer := binary.BigEndian.Uint32(c[n-4:])

	return topology.Layers[layer].LogicalMixes[posInLayer].Name
}

func (c Onion) ExtractCipher() *tibe.Cipher {
	cphr := &tibe.Cipher{}
	if err := cphr.SetBytes(c[:len(c)-8]); err != nil {
		panic("invalid onion")
	}

	return cphr
}

func (m MessageGenerator) onionWrap(
	round int, layer, posInLayer uint32, msg []byte, serverMasterPublicKey *tibe.MasterPublicKey) (Onion, error) {
	sname := m.SystemConfig.Topology.Layers[layer].LogicalMixes[posInLayer].Hostname

	id := computeId(sname, round)

	cipher, err := serverMasterPublicKey.Encrypt(id[:], msg)
	if err != nil {
		return Onion{}, err
	}

	bf := bytes.NewBuffer(make([]byte, 0, len(msg)+cipher.Size()+4*2))
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

func (m *MessageGenerator) MakeMessagesForClients(numClients, round int) []Onion {
	// todo: attempt to load it from file.
	return m.generateOnions(numClients, round)
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
