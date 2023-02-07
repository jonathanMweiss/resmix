package resmix

import (
	"github.com/jonathanMweiss/resmix/config"
	"github.com/jonathanMweiss/resmix/internal/crypto/tibe"
	"github.com/jonathanMweiss/resmix/internal/msync"
)

type (
	hostname string
	mixName  string

	Round int
)

type Sender interface {
	// TODO. decide how to send messages to other mixes.
}

type RoundState struct {
	MixHandler MixHandler
	Sender
}

type recoveryScheme struct {
	newTopology *config.Topology

	newResponsibility map[mixName]hostname
	keys              map[mixName]tibe.Decrypter
}

type MixHandler interface {
	SetKeys(keys map[mixName]tibe.Decrypter)
	// UpdateMixes states a failure and adds information regarding the new topology, keys etc.
	UpdateMixes(recoveryScheme)
	// AddMessages adds messages to a LogicalMix.
	AddMessages(messages []*tibe.Cipher)
	// GetOutputs returns the result of processings of the messages.
	GetOutputs() []*tibe.Cipher
}

type server struct {
	Publisher      tibe.Publisher
	DecryptionNode tibe.VssIbeNode // responsible for reconstructing decryption keys, generating keys for the current round.

	Configurations *config.ServerConfig
	Topology       *config.Topology

	States msync.Map[Round, RoundState]

	Connections map[hostname]MixClient
}
