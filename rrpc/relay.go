package rrpc

import (
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

func (s *Server) RelayStream(server Relay_RelayStreamServer) error {
	peer, err := GetPeerFromContext(server.Context())
	if err != nil {
		return err
	}

	if peer == "" {
		return status.Error(codes.InvalidArgument, "Missing caller identity")
	}

	peerId, err := s.ServerNetwork.GetPublicKey(peer)
	if err != nil {
		return err
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
