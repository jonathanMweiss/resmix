package resmix

import (
	"context"
	"fmt"
	"strconv"

	"github.com/golang/protobuf/proto"
	"github.com/jonathanMweiss/resmix/config"
	"github.com/jonathanMweiss/resmix/internal/msync"
	"github.com/jonathanMweiss/resmix/rrpc"
	"golang.org/x/crypto/sha3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func NewMixServer(cnfgs *ResmixConfigs) (MixServer, error) {
	pb, err := cnfgs.CreateTIBEPublisher()
	if err != nil {
		return nil, err
	}

	nd, err := cnfgs.CreateTIBENode()
	if err != nil {
		return nil, err
	}

	return &server{
		Publisher:      pb,
		DecryptionNode: nd,
		Configurations: cnfgs,
		States:         msync.Map[Round, RoundState]{},
		Connections:    map[hostname]rrpc.ClientConn{},
	}, nil
}

func computeId(hostname string, round int) []byte {
	bts := sha3.Sum256([]byte(hostname + strconv.Itoa(round)))
	return bts[:]
}

func (s *server) NewRound(ctx context.Context, request *NewRoundRequest) (*NewRoundResponse, error) {
	if s.Connections == nil {
		return nil, status.Error(codes.FailedPrecondition, "no connections to other mixes")
	}

	workloadPerMix := map[mixName]int{}
	for mName, load := range request.MixIdsToExpectedWorkload {
		workloadPerMix[mixName(mName)] = int(load)
	}

	// propagate workload:
	for _, mix := range s.Configurations.Topology.Layers[0].LogicalMixes {
		workload, ok := workloadPerMix[mixName(mix.Name)]
		if !ok {
			return nil, status.Error(codes.FailedPrecondition, fmt.Sprintf("no workload for first layer mix %s", mix.Name))
		}

		curr := mix
		for curr != nil {
			workloadPerMix[mixName(curr.Name)] = workload
			curr = s.Configurations.Topology.Mixes[curr.Successors[0]]
		}
	}

	mixes := []*config.LogicalMix{}
	for _, mix := range s.Configurations.Topology.Mixes {
		if mix.Hostname != s.Configurations.Hostname {
			continue
		}

		mixes = append(mixes, proto.Clone(mix).(*config.LogicalMix))
	}

	roundID := computeId(s.Configurations.Hostname, int(request.Round))
	decrypter := s.DecryptionNode.Decrypter(roundID[:])

	info := RoundState{
		MixHandler: NewMixers(s.Configurations.Topology, mixes, decrypter, workloadPerMix),
	}

	s.States.Store(Round(request.Round), info)

	return &NewRoundResponse{}, nil
}

func (s *server) EndRound(ctx context.Context, request *EndRoundRequest) (*EndRoundResponse, error) {
	state, ok := s.States.LoadAndDelete(Round(request.Round))
	if !ok {
		return nil, status.Error(codes.NotFound, "round not found")
	}

	state.Close()

	return &EndRoundResponse{}, nil
}

func (s *server) AddMessages(ctx context.Context, request *AddMessagesRequest) (*AddMessagesResponse, error) {
	state, ok := s.States.Load(Round(request.Round))
	if !ok {
		return nil, status.Error(codes.NotFound, "round not found")
	}

	s.ValidateTotalWork(ctx)
	state.MixHandler.AddMessages(request.Messages)

	return &AddMessagesResponse{}, nil
}

func (s *server) ValidateTotalWork(ctx context.Context) {
	// todo validate amount of messages match with the uuid.
	return
}

func (s *server) GetCoordinator() rrpc.ServerCoordinator {
	return s.ServerCoordinator
}

func (s *server) Dial() error {
	for _, peer := range s.Configurations.ServerConfig.Peers {
		host := peer.Hostname

		c, err := rrpc.NewConnection(host, s.Configurations.RrpcConfigs)
		if err != nil {
			if err := s.Close(); err != nil {
				return fmt.Errorf("failed to connect, then failed to close connections: %w", err)
			}

			return err
		}

		s.Connections[hostname(host)] = c
	}
	// will set up connections to other mixes.

	return nil
}

func (s *server) Close() error {
	for _, conn := range s.Connections {
		if err := conn.Close(); err != nil {
			return err
		}
	}

	return nil
}
