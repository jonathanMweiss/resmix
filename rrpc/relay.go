package rrpc

import (
	"io"
	"time"
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
	//TODO implement me
	// the relay receives something, stores it in a map that i filters over, if it haven't received a result by X time it creates an atterstation against it and publishes the attestations
	// todo: verify this client hadn't already sent this parcel
	server.Context()
	for {

		// TODO:
		relayRequest, err := server.Recv()
		if err == io.EOF {
			return nil
		}
		_ = relayRequest
		//if err != nil {
		//	return status.Error(codes.Internal, err.Error())
		//}

		//s.relayTasks.Store(
		//	relayRequest.Request.Parcel.Note.Calluuid,
		//	timedParcel{
		//		Time:      time.Now(),
		//		suspect:   relayRequest.Request.Parcel.Note.ReceiverID,
		//		requester: relayRequest.Request.Parcel.Note.SenderID,
		//	},
		//)
		// relay receives, pushes it onto network immediatly.
		// in case it wasn't received - attests.

		// i want to push using the network, but where do i wait for the response for?
		// should the server pass everything to the network? don't think so.
		// server received something, stores it into its attestor goroutine, then pushes it to the next one stream.
		// how do i have connection to that?
		// when i'll dial to the other server, and create my relayConns to push onto.
		// when receiving, its in the direction back on that stream.

		// TODO: network holds connection back to relays? no. Network is used by the clients mainly for relay relayConns.
		// server has its own connection, and it dosn't share them with multiple clients on the same app.

	}
}
