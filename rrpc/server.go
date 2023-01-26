package rrpc

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"runtime"
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
type server_signables struct {
	MerkleCertifiable
	signatureDone chan error
}

type Server struct {
	Services
	ServerNetwork
	verifier     *MerkleCertVerifier
	signingQueue chan server_signables
	skey         crypto.PrivateKey

	// closing the server fields:
	*sync.WaitGroup
	Cancel context.CancelFunc
	context.Context
	gsrvr *grpc.Server
}

func (s *Server) Stop() {
	s.gsrvr.Stop()
	s.Cancel()
	s.WaitGroup.Wait()
}

func (s *Server) Serve(lis net.Listener) error {
	return s.gsrvr.Serve(lis)
}

func NewServerService(skey crypto.PrivateKey, s Services, network ServerNetwork) RrpcServer {
	cntx, cancelf := context.WithCancel(context.Background())
	gsrvr := grpc.NewServer()

	srvr := &Server{
		Services:      s,
		ServerNetwork: network,
		verifier:      NewVerifier(runtime.NumCPU()),
		signingQueue:  make(chan server_signables, 100),
		Context:       cntx,
		Cancel:        cancelf,
		WaitGroup:     &sync.WaitGroup{},
		skey:          skey,
		gsrvr:         gsrvr,
	}

	RegisterServerServer(gsrvr, srvr)
	RegisterRelayServer(gsrvr, srvr)

	srvr.WaitGroup.Add(2)

	go serverSigner(srvr)

	go relayStreamSetup(srvr)

	return srvr
}

func relayStreamSetup(srvr *Server) {
	defer srvr.WaitGroup.Done()

	incomingChan := srvr.ServerNetwork.Incoming()
	_ = incomingChan

	// create someone that waits on all channel of anything that comes back from the network ...
}

func serverSigner(srvr *Server) {
	func() {
		defer srvr.WaitGroup.Done()

		for {
			select {
			case <-srvr.Context.Done():
				return
			case signTask := <-srvr.signingQueue:
				// TODO: batch more than one.
				signTask.signatureDone <- merkleSign([]MerkleCertifiable{signTask}, srvr.skey)
				close(signTask.signatureDone)
			}
		}
	}()
}

func (s *Server) Close() {
	s.Cancel()
	s.WaitGroup.Wait()
}

// CallStream is the part in the server that handles incoming rRPC parcels, forwards it to the server's collector to handle.
func (s *Server) CallStream(server Server_CallStreamServer) error {
	//TODO implement me
	for {
		req, err := server.Recv()
		if err == io.EOF {
			return nil
		}

		if err != nil {
			return err
		}
		// TODO
		_ = req
	}
}

func createDecodeFunc(payload []byte) func(v interface{}) error {
	return func(v interface{}) error {
		return codec.Unmarshal(payload, v)
	}
}

// DirectCall uses streams to represent a cheaper unary RPC.
func (s *Server) DirectCall(server Server_DirectCallServer) error {
	ip, err := GetPeerFromContext(server.Context())
	if err != nil {
		return status.Error(codes.InvalidArgument, "couldn't extract ip from incoming context: "+err.Error())
	}

	if len(ip) == 0 {
		return status.Error(codes.InvalidArgument, "empty ip")
	}

	clientPkey, err := s.ServerNetwork.GetPublicKey(ip)
	if err != nil {
		return status.Error(codes.InvalidArgument, "rrpc.network error in stream boot: "+err.Error())
	}

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

		if err := s.verifier.Verify(request.Note.SenderID, (*senderNote)(request.Note)); err != nil {
			return status.Error(codes.InvalidArgument, err.Error())
		}

		serviceOut, serviceError := methodDesc.Handler(svc.server, server.Context(), createDecodeFunc(request.Payload))

		result, err := intoDirectCallResponse(serviceError, serviceOut)
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		if err := s.signNote(request.Note); err != nil {
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

// Attest is an RPC call that any other Server uses to attest against another Server.
func (s *Server) Attest(server Relay_AttestServer) error {
	//TODO implement me
	panic("implement me")
}

func (s *Server) SendProof(server Relay_SendProofServer) error {
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

func (s *Server) signNote(note *ExchangeNote) error {
	resp := make(chan error)

	select {
	case <-s.Context.Done():
		return s.Context.Err()
	case s.signingQueue <- server_signables{
		MerkleCertifiable: (*receiverNote)(note),
		signatureDone:     resp,
	}:
		return <-resp
	}
}
