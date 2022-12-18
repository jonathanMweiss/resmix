package rrpc

import (
	"bytes"
	"fmt"
	"runtime"
	"sync"

	"github.com/jonathanMweiss/resmix/internal/codec"
	"github.com/jonathanMweiss/resmix/internal/crypto"
	"golang.org/x/net/context"
	status2 "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

type RrpcServer interface {
	RelayServer
	ServerServer
}

type server_signables struct {
	MerkleCertifiable
	signatureDone chan error
}

type Server struct {
	Services
	NetData
	*MerkleCertVerifier
	signingQueue chan server_signables
	skey         crypto.PrivateKey

	// closing the server fields:
	*sync.WaitGroup
	Cancel context.CancelFunc
	context.Context
}

func NewServerService(skey crypto.PrivateKey, s Services, network NetData) RrpcServer {
	cntx, cancelf := context.WithCancel(context.Background())
	srvr := &Server{
		Services:           s,
		NetData:            network,
		MerkleCertVerifier: NewVerifier(runtime.NumCPU()),
		signingQueue:       make(chan server_signables, 100),
		Context:            cntx,
		Cancel:             cancelf,
		WaitGroup:          &sync.WaitGroup{},
		skey:               skey,
	}

	srvr.WaitGroup.Add(1)
	go serverSigner(srvr)
	return srvr
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

func (s Server) Close() {
	s.Cancel()
	s.WaitGroup.Wait()
}

// TODO
func (s Server) CallStream(server Server_CallStreamServer) error {
	//TODO implement me
	panic("implement me")
}

func createDecodeFunc(payload []byte) func(v interface{}) error {
	return func(v interface{}) error {
		return codec.Unmarshal(payload, v)
	}
}

// DirectCall uses streams to represent a cheaper unary RPC.
func (s Server) DirectCall(server Server_DirectCallServer) error {
	ip, err := GetPeerFromContext(server.Context())
	if err != nil {
		return status.Error(codes.InvalidArgument, "couldn't extract ip from incoming context: "+err.Error())
	}

	if len(ip) == 0 {
		return status.Error(codes.InvalidArgument, "empty ip")
	}

	clientPkey, err := s.NetData.GetPublicKey(ip)
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

		if err := s.Verify(request.Note.SenderID, (*senderNote)(request.Note)); err != nil {
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

func (s Server) RelayStream(server Relay_RelayStreamServer) error {
	//TODO implement me
	panic("implement me")
}

func (s Server) Attest(server Relay_AttestServer) error {
	//TODO implement me
	panic("implement me")
}

func (s Server) SendProof(server Relay_SendProofServer) error {
	//TODO implement me
	panic("implement me")
}

func (s Server) signNote(note *ExchangeNote) error {
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
