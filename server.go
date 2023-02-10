package resmix

import (
	"context"
	"github.com/golang/protobuf/proto"
	"github.com/jonathanMweiss/resmix/config"
	"github.com/jonathanMweiss/resmix/internal/crypto/tibe"
	"github.com/jonathanMweiss/resmix/internal/msync"
	"github.com/jonathanMweiss/resmix/rrpc"
	"golang.org/x/crypto/sha3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strconv"
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
		MixHandler: NewMixers(s.Configurations.Hostname, mixes, decrypter, workloadPerMix),
	}

	s.States.Store(Round(request.Round), info)

	return &NewRoundResponse{}, nil
}

func NewMixers(hostname string, mixes []*config.LogicalMix, workloadMap tibe.Decrypter, decrypter map[mixName]int) MixHandler {
	return nil
}

func (s *server) EndRound(ctx context.Context, request *EndRoundRequest) (*EndRoundResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s *server) AddMessages(ctx context.Context, request *AddMessagesRequest) (*AddMessagesResponse, error) {
	//TODO implement me
	panic("implement me")
}
