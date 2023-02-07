package resmix

import (
	"context"
	"github.com/golang/protobuf/proto"
	"github.com/jonathanMweiss/resmix/config"
	"github.com/jonathanMweiss/resmix/internal/crypto/tibe"
	"golang.org/x/crypto/sha3"
	"strconv"
)

func computeId(hostname string, round int) [32]byte {
	return sha3.Sum256([]byte(hostname + strconv.Itoa(round)))
}

func (s *server) NewRound(ctx context.Context, request *NewRoundRequest) (*NewRoundResponse, error) {
	workloadPerMix := map[mixName]int{}
	for mName, load := range request.MixIdsToExpectedWorkload {
		workloadPerMix[mixName(mName)] = int(load)
	}

	mixes := []*config.LogicalMix{}
	for _, mix := range s.Topology.Mixes {
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
