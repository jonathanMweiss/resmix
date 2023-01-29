package rrpc

import (
	"fmt"
	"io"
	"time"

	"github.com/jonathanMweiss/resmix/internal/crypto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TODO: attestor, should have a priority queue of items to be ready to attest against.
// runs on ticker of 200ms to check the top,
// on tick will check if it's time to attest on the top. if it is pops it and checks on the next one too.

type timedParcel struct {
	time.Time
	// must answer before this timed parcel is deleted
	suspect   []byte
	requester []byte
}

func relayStreamSetup(srvr *Server) {
	defer srvr.WaitGroup.Done()

	// TODO: map uuid to client name.
	// TODO: map client name to stream object.

	incomingChan := srvr.ServerNetwork.Incoming()
	for callStreamResponse := range incomingChan {
		fmt.Println("got a call stream response")
		// todo: check if i know this request.
		// attempt to restore request to the original requester (push the result onto `server Relay_RelayStreamServer`)
		_ = callStreamResponse
	}

	// create someone that waits on all channel of anything that comes back from the network ...
}

func (s *Server) RelayStream(server Relay_RelayStreamServer) error {
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

		s.ServerNetwork.AsyncSend(
			relayRequest.Request.Parcel.Note.ReceiverID,
			&CallStreamRequest{
				Parcel: relayRequest.Request.Parcel,
			},
		)
	}
}

func isValidRequest(id crypto.PublicKey, request *RelayStreamRequest) bool {
	return true // todo.

}

func (s *Server) logRequestAsReceived(request *RelayStreamRequest) {
	// todo
}
