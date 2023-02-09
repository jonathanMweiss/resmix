package rrpc

import (
	"fmt"
	"github.com/jonathanMweiss/resmix/internal/crypto"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"runtime"

	"github.com/jonathanMweiss/resmix/internal/ecc"
)

// Coordinator contains any data used in the rrpc, along with Connections to the relays.
type Coordinator interface {
	NetworkData

	GetSecretKey() crypto.PrivateKey

	getErrorCorrectionCode() ecc.VerifyingEncoderDecoder
	getVerifier() *MerkleCertVerifier
	getRelayGroup() RelayGroup

	Dial() error
	Close() error
}

type RelayGroup interface {
	CancelRequest(uuid string)
	RobustRequest(context context.Context, requests []*RelayRequest) ([]*CallStreamResponse, error)
	PublishProof(*Proof)
}

// ServerCoordinator is an advanced Coordinator, used by relays to utilise the open relay connections in a coordinator
// to receive incoming data, and communicate with other relays.
type ServerCoordinator interface {
	Coordinator

	AsyncSend(publickey crypto.PublicKey, msg *CallStreamRequest) error

	// Incoming returns anything(!) even timeouts that came over the
	// coordinator, from any of the servers this coordinator is listening on!
	Incoming() <-chan *CallStreamResponse
}

// coordinator implements the ServerCoordinator interface
type coordinator struct {
	NetworkData
	*MerkleCertVerifier

	skey        crypto.PrivateKey
	relayConns  map[string]*RelayConn
	serverConns map[string]*ServerConn

	callResponseChan chan *CallStreamResponse

	myAddress string

	ctx    context.Context
	cancel context.CancelFunc
	ecc    ecc.VerifyingEncoderDecoder
}

func (n *coordinator) GetSecretKey() crypto.PrivateKey {
	return n.skey
}

// NewCoordinator creates a Coordinator that is tied to a speicific node. cannot reuse for different nodes on same machine!
func NewCoordinator(netdata NetworkData, skey crypto.PrivateKey) ServerCoordinator {
	maxErasures := netdata.MaxErrors()
	numDataShards := len(netdata.Servers()) - maxErasures

	ecc_, err := ecc.NewRSEncoderDecoder(numDataShards, maxErasures)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	v := NewVerifier(runtime.NumCPU())
	myAddress := netdata.GetHostname(skey.Public())
	ctx = addIPToContext(ctx, myAddress)

	return &coordinator{
		NetworkData: netdata,
		skey:        skey,

		relayConns:  make(map[string]*RelayConn, len(netdata.Servers())),
		serverConns: make(map[string]*ServerConn, len(netdata.Servers())),

		MerkleCertVerifier: v,
		ecc:                ecc_,

		ctx:    ctx,
		cancel: cancel,

		myAddress: myAddress,

		callResponseChan: make(chan *CallStreamResponse, 100),
	}
}

// functionality:

func (n *coordinator) Close() error {
	defer n.MerkleCertVerifier.Stop()
	defer n.ecc.Stop()

	close(n.callResponseChan)

	n.cancel()

	var err error
	for _, conn := range n.relayConns {
		if err = conn.Close(); err != nil {
			fmt.Println("error closing relay connection:", err)
		}
	}

	for _, conn := range n.serverConns {
		if err = conn.Close(); err != nil {
			fmt.Println("error closing server connection:", err)
		}
	}

	return err
}

func (n *coordinator) getErrorCorrectionCode() ecc.VerifyingEncoderDecoder {
	return n.ecc
}

func (n *coordinator) getVerifier() *MerkleCertVerifier {
	return n.MerkleCertVerifier
}

func (n *coordinator) AsyncSend(publickey crypto.PublicKey, msg *CallStreamRequest) error {
	hostname := n.GetHostname(publickey)
	if hostname == "" {
		return status.Error(codes.NotFound, "public key not found")
	}

	conn, ok := n.serverConns[hostname]
	if !ok {
		return status.Error(codes.Internal, "no connection to server")
	}

	conn.send(msg)
	return nil
}

func (n *coordinator) Incoming() <-chan *CallStreamResponse {
	return n.callResponseChan
}

func (n *coordinator) getRelayGroup() RelayGroup {
	return n
}

func (n *coordinator) RobustRequest(ctx context.Context, requests []*RelayRequest) ([]*CallStreamResponse, error) {
	if len(requests) != len(n.relayConns) {
		return nil, fmt.Errorf("bad request, number of requests differs from number of relays")
	}

	serverID := requests[0].Parcel.Note.ReceiverID
	responseChan := make(chan relayResponse, len(requests))
	srvrs := n.Servers()

	for i := range requests {
		v, ok := n.relayConns[srvrs[requests[i].Parcel.RelayIndex]]
		if !ok {
			panic("relay index not found")
		}

		v.sendRequest(relayConnRequest{
			RelayRequest: requests[i],
			response:     responseChan,
		})
	}

	responses := make([]*CallStreamResponse, 0, n.MinimalRelayedParcels())
	totalErrors := 0

	var err error

	for {
		select {
		case <-n.ctx.Done():
			return nil, n.ctx.Err()
		case <-ctx.Done():
			return nil, ctx.Err()
		case r := <-responseChan:
			if tmperr := n.validateRrpcResponse(serverID, r); tmperr != nil {
				totalErrors += 1

				if totalErrors > n.MaxErrors() {
					err = tmperr

					return nil, err
				}

				continue
			}

			responses = append(responses, r.Response)

			if len(responses) >= n.MinimalRelayedParcels() {
				return responses, nil
			}
		}

	}
}

func (n *coordinator) validateRrpcResponse(serverId crypto.PublicKey, r relayResponse) error {
	switch {
	case r.RelayStreamError != nil:
		return status.ErrorProto(r.RelayStreamError)

	case r.Response == nil:
		return status.Error(codes.Internal, "empty response")

	case r.Response.RpcError != nil:
		return status.ErrorProto(r.Response.RpcError)

	}

	return n.MerkleCertVerifier.Verify(serverId, r.Response)
}

func (n *coordinator) CancelRequest(uuid string) {
	for _, conn := range n.relayConns {
		conn.cancelRequest(uuid)
	}
}

func (n *coordinator) PublishProof(p *Proof) {
	for _, conn := range n.relayConns {
		conn.SendProof(p)
	}
}

func (n *coordinator) GetRelayConn(hostname string) *RelayConn {
	return n.relayConns[hostname]
}

func (n *coordinator) Dial() error {
	for index, s := range n.NetworkData.Servers() {
		relayConn, err := newRelayConn(n.ctx, s, index)
		if err != nil {
			return n.Close()
		}

		n.relayConns[s] = relayConn

		serverConn, err := newServerConn(n.ctx, s, n.callResponseChan)
		if err != nil {
			return n.Close()
		}
		n.serverConns[s] = serverConn
	}

	return nil
}
