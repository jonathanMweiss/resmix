package resmix

import (
	"github.com/jonathanMweiss/resmix/config"
	"github.com/jonathanMweiss/resmix/internal/crypto/tibe"
	"sync"
)

type Mixers struct {
	Mixes map[mixName]Mix
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

func NewMixers(hostname string, mixes []*config.LogicalMix, decrypter tibe.Decrypter, workloadMap map[mixName]int) MixHandler {

	// advance map of workload such that all mixes inherit the work of their predecessors... (starting with the first layer onward)
	return nil
}

type decryptionTask struct {
	wg *sync.WaitGroup

	tibe.Decrypter
	input Onion

	// output onion on the
	Idx    int
	Result []Onion
}

func (m Mixers) SetKeys(keys map[mixName]tibe.Decrypter) {
	//TODO implement me
	panic("implement me")
}

func (m Mixers) UpdateMixes(scheme recoveryScheme) {
	//TODO implement me
	panic("implement me")
}

func (m Mixers) AddMessages(messages []*Messages) {
	//TODO implement me
	panic("implement me")
}

func (m Mixers) GetOutputs() []*tibe.Cipher {
	//TODO implement me
	panic("implement me")
}
