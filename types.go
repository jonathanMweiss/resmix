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
	Close() error
	GetCoordinator() rrpc.ServerCoordinator
}
type (
	hostname string
	mixName  string

	Round int
)

type Sender interface {
	Close()
	// TODO. decide how to send messages to other mixes.
}

type RoundState struct {
	MixHandler MixHandler
	Sender
}

func (s RoundState) Close() {
	s.MixHandler.Close()
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

type MixHandler interface {
	// UpdateMixes states a failure and adds information regarding the new topology, keys etc.
	UpdateMixes(recoveryScheme)
	// AddMessages adds messages to a LogicalMix.
	AddMessages(messages []*Messages)
	// GetOutputs returns the result of processings of the messages.
	GetOutputsChan() <-chan mixoutput

	Close()
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

	Connections map[hostname]rrpc.ClientConn

	rrpc.ServerCoordinator
}
