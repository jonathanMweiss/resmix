package rrpc

import (
	"context"
	"fmt"
	"github.com/jonathanMweiss/resmix/internal/crypto"
	"github.com/jonathanMweiss/resmix/internal/msync"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"sync"
	"time"
)

// TODO: attestor, should have a priority queue of items to be ready to attest against.
// 	uses heap, always sorts them according to time, updates its ticker according to top.
//	either a new item is received, or the time pops..

type relaytask struct {
	uuid      string
	startTime time.Time
	respchan  chan *RelayStreamResponse
}

func (r *relaytask) GetStartTime() time.Time {
	return r.startTime
}

func (r *relaytask) PrepareForDeletion() {
	r.respchan <- &RelayStreamResponse{
		RelayStreamError: status.New(codes.Canceled, "relay: task timed out").Proto(),
		Uuid:             r.uuid,
	}
}

type relay struct {
	*sync.WaitGroup
	context.Context
	ServerNetwork ServerNetwork

	uuidToPeer msync.Map[string, relaytask]
}

func (s *relay) SendProof(server Relay_SendProofServer) error {
	// TODO need to receive some kind of Attestor that once started will look for specific uuids and their items!

	server.Context() // todo something with context of client. like verify it is a known client
	for {
		r, err := server.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		_ = r
	}
	//panic("implement me")
}

func (s *relay) Attest(server Relay_AttestServer) error {
	//TODO implement me
	panic("implement me")
}

func (srvr *relay) relaySetup() {
	srvr.WaitGroup.Add(2)

	go func() {
		defer srvr.WaitGroup.Done()

		incomingChan := srvr.ServerNetwork.Incoming()
		for {
			select {
			case <-srvr.Context.Done():
				return

			case c, ok := <-incomingChan:
				if !ok {
					return
				}
				fmt.Println("got a call stream response", c.String())
			}
		}
	}()

	go func() {
		defer srvr.WaitGroup.Done()

		foreverCleanup(srvr.Context, &srvr.uuidToPeer)
	}()
}

func (s *relay) RelayStream(server Relay_RelayStreamServer) error {
	peer, err := GetPeerFromContext(server.Context())
	if err != nil {
		return status.Error(codes.Unauthenticated, "server: cannot get peer from context")

	}

	peerId, err := s.ServerNetwork.GetPublicKey(peer)
	if err != nil {
		return status.Errorf(codes.Unauthenticated, "server: cannot get peer from context %v", err)
	}

	for {
		relayRequest, err := server.Recv()
		if err == io.EOF {
			return nil
		}

		if err != nil {
			return err
		}

		if !isValidRequest(peerId, relayRequest) {
			continue
		}

		s.logRequestAsReceived(relayRequest)

		err = s.ServerNetwork.AsyncSend(
			relayRequest.Request.Parcel.Note.ReceiverID,
			&CallStreamRequest{
				Parcel: relayRequest.Request.Parcel,
			},
		)
		if err != nil {
			continue
		}
	}
}

func isValidRequest(id crypto.PublicKey, request *RelayStreamRequest) bool {
	return true // todo.

}

func (s *relay) logRequestAsReceived(request *RelayStreamRequest) {
	// todo
}
