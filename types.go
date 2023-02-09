package resmix

import (
	"github.com/jonathanMweiss/resmix/config"
	"github.com/jonathanMweiss/resmix/internal/crypto/tibe"
	"github.com/jonathanMweiss/resmix/internal/msync"
	"github.com/jonathanMweiss/resmix/rrpc"
)

type ResMixServer interface {
	MixServer

	// Dial is used to connect to other mixes.
	Dial() error
	GetCoordinator() rrpc.ServerCoordinator
}
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
type ResmixConfigs struct {
	*config.ServerConfig
	*config.Topology
	RrpcConfigs rrpc.Configs
}

type server struct {
	Publisher      tibe.Publisher
	DecryptionNode tibe.VssIbeNode // responsible for reconstructing decryption keys, generating keys for the current round.

	Configurations *ResmixConfigs

	States msync.Map[Round, RoundState]

	Connections map[hostname]MixClient

	rrpc.ServerCoordinator
}

func (s *server) GetCoordinator() rrpc.ServerCoordinator {
	return s.ServerCoordinator
}

func (s *server) Dial() error {
	// will set up connections to other mixes.

	return nil
}
