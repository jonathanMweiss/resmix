package rrpc

import (
	"context"
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
	ServerNetwork ServerCoordinator

	uuidToPeer msync.Map[string, relaytask]
}

func (r *relay) SendProof(server Relay_SendProofServer) error {
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

func (r *relay) Attest(server Relay_AttestServer) error {
	//TODO implement me
	panic("implement me")
}

func (r *relay) relaySetup() {
	r.WaitGroup.Add(2)

	go func() {
		defer r.WaitGroup.Done()

		incomingChan := r.ServerNetwork.Incoming()
		for {
			select {
			case <-r.Context.Done():
				return

			case c, ok := <-incomingChan:
				if !ok {
					return
				}

				rqst, exists := r.uuidToPeer.LoadAndDelete(c.Note.Calluuid)
				if !exists {
					continue
				}

				rqst.respchan <- &RelayStreamResponse{
					Response: c,
					Uuid:     c.Note.Calluuid,
				}
			}
		}
	}()

	go func() {
		defer r.WaitGroup.Done()

		foreverCleanup(r.Context, &r.uuidToPeer)
	}()
}

func (r *relay) RelayStream(server Relay_RelayStreamServer) error {
	peer, err := getPeerFromContext(server.Context())
	if err != nil {
		return status.Error(codes.Unauthenticated, "server: cannot get peer from context")

	}

	peerId, err := r.ServerNetwork.GetPublicKey(peer)
	if err != nil {
		return status.Errorf(codes.Unauthenticated, "server: cannot get peer from context %v", err)
	}
	resps := make(chan *RelayStreamResponse, 100)
	go func() {
		for r := range resps {
			if err := server.Send(r); err != nil {
				return
			}
		}
	}()

	for {
		relayRequest, err := server.Recv()
		if err == io.EOF {
			return nil
		}

		if err != nil {
			return err
		}

		if !isValidRequest(peerId, relayRequest) {
			resps <- &RelayStreamResponse{
				RelayStreamError: status.New(codes.InvalidArgument, "relay: invalid request").Proto(),
				Uuid:             relayRequest.Request.Parcel.Note.Calluuid,
			}
		}

		r.logRequestAsReceived(relayRequest)

		err = r.ServerNetwork.AsyncSend(
			relayRequest.Request.Parcel.Note.ReceiverID,
			&CallStreamRequest{
				Parcel: relayRequest.Request.Parcel,
			},
		)

		if err != nil {
			st, _ := status.FromError(err)
			resps <- &RelayStreamResponse{
				RelayStreamError: st.Proto(),
				Uuid:             relayRequest.Request.Parcel.Note.Calluuid,
			}

			continue
		}

		r.uuidToPeer.Store(relayRequest.Request.Parcel.Note.Calluuid, relaytask{
			uuid:      relayRequest.Request.Parcel.Note.Calluuid,
			startTime: time.Now(),
			respchan:  resps,
		})
	}
}

func isValidRequest(id crypto.PublicKey, request *RelayStreamRequest) bool {
	return true // todo.

}

func (r *relay) logRequestAsReceived(request *RelayStreamRequest) {
	// todo
}
