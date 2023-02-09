package rrpc

import (
	"bytes"
	"fmt"
	"net"
	"sync"

	"github.com/jonathanMweiss/resmix/internal/codec"
	"github.com/jonathanMweiss/resmix/internal/crypto"
	"golang.org/x/net/context"
	status2 "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

type RrpcServer interface {
	RelayServer
	ServerServer
	Serve(lis net.Listener) error
	Stop()
}

type server struct {
	Services
	ServerCoordinator

	skey crypto.PrivateKey

	collectorTasks chan *Parcel

	// closing the server fields:
	*sync.WaitGroup
	context.CancelFunc
	context.Context
	gsrvr *grpc.Server

	streamsBack srvrStreams

	bufferPool sync.Pool

	*relay
}

func (s *server) Stop() {
	s.gsrvr.Stop()
	s.CancelFunc()
	s.WaitGroup.Wait()
}

func (s *server) Serve(lis net.Listener) error {
	return s.gsrvr.Serve(lis)
}

func newServerService(s Services, network ServerCoordinator, options ...grpc.ServerOption) (RrpcServer, error) {
	cntx, cancelf := context.WithCancel(context.Background())
	gsrvr := grpc.NewServer(options...)

	srvr := &server{
		Services:          s,
		ServerCoordinator: network,

		skey: network.GetSecretKey(),

		WaitGroup:  &sync.WaitGroup{},
		CancelFunc: cancelf,
		Context:    cntx,
		gsrvr:      gsrvr,

		collectorTasks: make(chan *Parcel, 1000),

		streamsBack: newStreams(network),

		bufferPool: sync.Pool{New: func() interface{} { return bytes.NewBuffer(make([]byte, 0, 1024)) }},
	}
	srvr.PrepareRelayService()

	srvr.RegIntoGrpc()

	srvr.relay.relaySetup()

	srvr.WaitGroup.Add(1)
	go srvr.collector()

	return srvr, nil
}

func createDecodeFunc(payload []byte) func(v interface{}) error {
	return func(v interface{}) error {
		return codec.Unmarshal(payload, v)
	}
}

// DirectCall uses streams to represent a cheaper unary RPC.
func (s *server) DirectCall(server Server_DirectCallServer) error {
	ip, err := getPeerFromContext(server.Context())
	if err != nil {
		return status.Errorf(codes.Unauthenticated, "server::dirceCall: cannot get peer from context: %v", err)
	}

	clientPkey, err := s.ServerCoordinator.GetPublicKey(ip)
	if err != nil {
		return status.Errorf(codes.Unauthenticated, "server::dirceCall: unknown caller: %v", err)
	}

	verifier := s.ServerCoordinator.getVerifier()
	for {
		request, err := server.Recv()
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		svc, methodDesc, err := s.Services.getServiceAndMethodDesc(request.Method)
		if err != nil {
			return status.Error(codes.Unknown, err.Error())
		}

		if request.Note == nil {
			return status.Error(codes.InvalidArgument, "missing note")
		}

		if !bytes.Equal(clientPkey, request.Note.SenderID) {
			return status.Error(
				codes.InvalidArgument,
				"received exchange note with sender id not matching the client's public key",
			)
		}

		if err := verifier.Verify(request.Note.SenderID, (*senderNote)(request.Note)); err != nil {
			return status.Error(codes.InvalidArgument, err.Error())
		}

		serviceOut, serviceError := methodDesc.Handler(svc.Server, server.Context(), createDecodeFunc(request.Payload))

		result, err := intoDirectCallResponse(serviceError, serviceOut)
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		if err := merkleSign([]MerkleCertifiable{(*receiverNote)(request.Note)}, s.skey); err != nil {
			fmt.Println("couldn't sign the note. exiting stream:", err.Error())
			return status.Error(codes.Internal, err.Error())
		}

		response := DirectCallResponse{
			Note:   request.Note,
			Result: result,
		}

		err = server.Send(&response)
		if err != nil {
			fmt.Println("server faced an error while sending a response: ", err.Error())
			return status.Error(codes.Internal, err.Error())
		}
	}
}

func intoDirectCallResponse(serviceError error, serviceOut interface{}) (isDirectCallResponse_Result, error) {
	if serviceError != nil {
		return &DirectCallResponse_RpcError{errToRpcErr(serviceError)}, nil
	}

	responsePayload, err := codec.Marshal(serviceOut)
	if err != nil {
		return nil, err
	}

	return &DirectCallResponse_Payload{responsePayload}, nil
}

func errToRpcErr(serviceError error) *status2.Status {
	err, ok := status.FromError(serviceError)
	if !ok {
		return status.New(codes.Unknown, serviceError.Error()).Proto()
	}

	return err.Proto()
}

func (s *server) PrepareRelayService() {
	s.relay = &relay{
		WaitGroup:     s.WaitGroup,
		Context:       s.Context,
		ServerNetwork: s.ServerCoordinator,
	}
}

func (s *server) RegIntoGrpc() {
	RegisterServerServer(s.gsrvr, s)
	RegisterRelayServer(s.gsrvr, s.relay)
}
