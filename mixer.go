package resmix

import (
	"sync"

	"github.com/jonathanMweiss/resmix/config"
	"github.com/jonathanMweiss/resmix/internal/crypto/tibe"
)

type Mixers struct {
	Mixes map[mixName]*Mix
}

type physyicalMixName struct {
	Hostname hostname
	MixName  mixName
}

type Mix struct {
	decryptionKey tibe.Decrypter

	predecessors  map[physyicalMixName]int // usually one, unless the host of that mix failed.
	successors    []mixName                // usually one, unless the host of that mix failed.
	totalReceived map[physyicalMixName]int // total number of messages received from each predecessor.

	totalWorkload int
	outputs       []Onion // should be of size totalWorkload

	wg sync.WaitGroup // used to wait for all messages to be processed by the threadpool.
}

func NewMixers(topo *config.Topology, mixesConfigs []*config.LogicalMix, decrypter tibe.Decrypter, workloadMap map[mixName]int) MixHandler {
	Mixes := make(map[mixName]*Mix)

	for _, mix := range mixesConfigs {
		m := &Mix{
			decryptionKey: decrypter,
			predecessors:  make(map[physyicalMixName]int),
			successors:    make([]mixName, len(mix.Successors)),
			totalReceived: make(map[physyicalMixName]int),
			totalWorkload: workloadMap[mixName(mix.Name)],
			outputs:       make([]Onion, workloadMap[mixName(mix.Name)]),
		}

		for i, successor := range m.successors {
			m.successors[i] = successor
		}

		for _, predecessor := range mix.Predecessors {
			host := hostname(config.GenesisName)
			if predecessor != config.GenesisName {
				host = hostname(topo.Mixes[predecessor].Hostname)
			}

			pname := physyicalMixName{
				Hostname: host,
				MixName:  mixName(predecessor),
			}

			m.predecessors[pname] = workloadMap[mixName(predecessor)]
		}

		for name := range m.predecessors {
			m.totalReceived[name] = 0
		}

		Mixes[mixName(mix.Name)] = m
	}

	m := &Mixers{
		Mixes: Mixes,
	}

	return m
}

type decryptionTask struct {
	wg *sync.WaitGroup

	tibe.Decrypter
	input Onion

	// output onion on the
	Idx    int
	Result []Onion
}

func (m *Mixers) SetKeys(keys map[mixName]tibe.Decrypter) {
	//TODO implement me
	panic("implement me")
}

func (m *Mixers) UpdateMixes(scheme recoveryScheme) {
	//TODO implement me
	panic("implement me")
}

func (m *Mixers) AddMessages(messages []*Messages) {
	//TODO implement me
	panic("implement me")
}

func (m *Mixers) GetOutputs() []*tibe.Cipher {
	//TODO implement me
	panic("implement me")
}
