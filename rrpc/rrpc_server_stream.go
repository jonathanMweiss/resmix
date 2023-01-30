package rrpc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/jonathanMweiss/resmix/internal/codec"
	"io"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type srvrStreams struct {
	chans []chan *CallStreamResponse
}

func newStreams(serverNetwork ServerNetwork) srvrStreams {
	chans := make([]chan *CallStreamResponse, len(serverNetwork.Servers()))
	for i := range chans {
		chans[i] = make(chan *CallStreamResponse, 100)
	}

	return srvrStreams{
		chans: chans,
	}
}

func (s *srvrStreams) sendTo(index int, resp *CallStreamResponse) {
	s.chans[index] <- resp
}

func (s *srvrStreams) getChan(index int) <-chan *CallStreamResponse {
	return s.chans[index]
}

func (s *srvrStreams) removeAt(index int) {
	close(s.chans[index])
}

// CallStream is the part in the server that handles incoming rRPC parcels, forwards it to the server's collector to handle.
func (s *Server) CallStream(stream Server_CallStreamServer) error {
	peerIp, err := GetPeerFromContext(stream.Context())
	if err != nil {
		return status.Errorf(codes.Unauthenticated, "server::callStream: cannot get peer from context: %v", err)
	}

	// verify existence of the caller.
	if _, err := s.ServerNetwork.GetPublicKey(peerIp); err != nil {
		return status.Errorf(codes.Unauthenticated, "server::callStream: unknown caller: %v", err)
	}

	relayIndex := s.ServerNetwork.GetRelayIndex(peerIp)

	// used by the server to send messages to the relay.
	go func() {
		for tosend := range s.streamsBack.getChan(relayIndex) {
			if err := stream.Send(tosend); err != nil {
				s.streamsBack.removeAt(relayIndex)
				return
			}
		}
	}()

	for {
		request, err := stream.Recv()
		if err == io.EOF {
			return nil
		}

		if err != nil {
			return err
		}

		if err := s.validateParcel(relayIndex, request.Parcel); err != nil {
			s.streamsBack.sendTo(relayIndex,
				&CallStreamResponse{
					RpcError: status.Newf(codes.InvalidArgument, "server:"+err.Error()).Proto(),
				})

			continue
		}

		select {
		case <-s.Context.Done():
			return status.Error(codes.Canceled, "server: stopping operation")
		case s.collectorTasks <- request.Parcel:
		}
		// move for job collection.
	}
}

func (s *Server) validateParcel(index int, parcel *Parcel) error {
	if int(parcel.RelayIndex) != index {
		return status.Errorf(codes.InvalidArgument, "server: invalid relay index: %d", parcel.RelayIndex)
	}

	tmpCert := (*senderNote)(parcel.Note).popCert()
	if err := s.ServerNetwork.getVerifier().Verify(parcel.Note.SenderID, parcel); err != nil {
		return err
	}
	(*senderNote)(parcel.Note).pushCert(tmpCert)

	// verify the note contains a valid signature too. Otherwise, you cannot fill it!
	return s.ServerNetwork.getVerifier().Verify(parcel.Note.ReceiverID, (*senderNote)(parcel.Note))
}

type rrpcTask struct {
	parcels   []*Parcel
	savedNote *ExchangeNote
	startTime time.Time
	service   *ServiceDesc
	method    *MethodDesc
}

// collector reduces the incoming parcels into a task.
// once we reach enough parcels, we can decode them and run them.
func (s *Server) collector() {
	defer s.WaitGroup.Done()

	ttl := time.Second * 5
	timeToLiveTicker := time.NewTicker(ttl)

	tasks := map[string]*rrpcTask{}

	for {
		select {
		case <-timeToLiveTicker.C:
			for k, v := range tasks {
				if time.Since(v.startTime) > ttl {
					delete(tasks, k)
				}
			}

		case <-s.Context.Done():
			return

		case task := <-s.collectorTasks:
			v, err := s.getOrCreateTask(tasks, task)

			if err != nil {
				s.streamsBack.sendTo(int(task.RelayIndex), &CallStreamResponse{
					RpcError: status.Newf(codes.InvalidArgument, "server:"+err.Error()).Proto(),
				})

				continue
			}

			v.parcels = append(v.parcels, task)
			if len(v.parcels) < s.ServerNetwork.MinimalRelayedParcels() {
				continue
			}

			delete(tasks, v.savedNote.Calluuid)

			// todo: consider using a thread pool...
			go func() {
				resps := s.runTask(v)
				signables := make([]MerkleCertifiable, len(resps))

				for i, resp := range resps {
					signables[i] = resp
				}

				if err := merkleSign(signables, s.skey); err != nil {
					fmt.Println("server: couldn't sign rrpc response", err)
					return
				}

				for i, resp := range resps {
					s.streamsBack.sendTo(i, resp)
				}
			}()

		}
	}
}

func (s *Server) getOrCreateTask(tasks map[string]*rrpcTask, parcel *Parcel) (*rrpcTask, error) {
	service, methodDesc, err := s.getServiceAndMethodDesc(parcel.Method)
	if err != nil {
		return nil, err
	}

	v, ok := tasks[parcel.Note.Calluuid]
	if !ok {
		v = &rrpcTask{
			parcels:   make([]*Parcel, 0, s.ServerNetwork.MinimalRelayedParcels()),
			savedNote: parcel.Note,
			startTime: time.Now(),
			service:   service,
			method:    methodDesc,
		}

		tasks[parcel.Note.Calluuid] = v
	}

	return v, nil
}

func (s *Server) runTask(v *rrpcTask) []*CallStreamResponse {
	payload, err := s.reconstructParcels(v)
	if err != nil {
		return s.erroToCallStreamResponseArray(v, err)
	}

	resp, err := v.method.Handler(v.service.server, s.Context, createDecodeFunc(payload))
	if err != nil {
		return s.erroToCallStreamResponseArray(v, err)
	}

	resps, err := s.prepareCallResponse(resp, v)
	if err != nil {
		return s.erroToCallStreamResponseArray(v, err)
	}

	return resps
}

func (s *Server) reconstructParcels(v *rrpcTask) ([]byte, error) {
	if len(v.parcels) == 0 {
		return nil, status.Error(codes.Internal, "reached reconstruction with 0 parcels")
	}

	msgSize := binary.LittleEndian.Uint32(v.parcels[0].MessageLength)

	shards := s.ServerNetwork.getErrorCorrectionCode().NewShards()

	for i := 0; i < len(v.parcels); i++ {
		(*eccClientParcel)(v.parcels[i]).PutIntoShards(shards)

		v.parcels[i] = nil
	}

	data, err := s.ServerNetwork.getErrorCorrectionCode().AuthReconstruct(shards, int(msgSize))
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "server reconstruction failure: %v", err)
	}

	return data, nil
}

func (s *Server) prepareCallResponse(response interface{}, v *rrpcTask) ([]*CallStreamResponse, error) {
	bf := s.bufferPool.Get().(*bytes.Buffer)
	defer s.bufferPool.Put(bf)

	if err := codec.MarshalIntoWriter(response, bf); err != nil {
		return nil, err
	}

	msgLength := make([]byte, 4)
	binary.LittleEndian.PutUint32(msgLength, uint32(len(bf.Bytes())))

	chunks, err := s.ServerNetwork.getErrorCorrectionCode().AuthEncode(bf.Bytes())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failure in encoding serverside, %v:", err)
	}

	toRelayBack := make([]*CallStreamResponse, len(s.ServerNetwork.Servers()))
	for i := range s.ServerNetwork.Servers() {
		srsp := &CallStreamResponse{
			Response: &RrpcResponse{
				MessageLength: msgLength,
				RelayIndex:    int32(i),
			},
			PublicKey: s.skey.Public(),
			Note:      v.savedNote,
		}

		(*eccServerParcel)(srsp.Response).InsertECCPayload(chunks, i)
		toRelayBack[i] = srsp
	}

	return toRelayBack, nil
}

func (s *Server) erroToCallStreamResponseArray(v *rrpcTask, err error) []*CallStreamResponse {
	st, ok := status.FromError(err)
	if !ok {
		st = status.New(codes.Unknown, err.Error())
	}
	statusErrorProto := st.Proto()

	sresponse := make([]*CallStreamResponse, len(s.ServerNetwork.Servers()))
	for i := range s.ServerNetwork.Servers() {
		sresponse[i] = &CallStreamResponse{
			RpcError:  statusErrorProto,
			PublicKey: s.skey.Public(),
			Note: &ExchangeNote{
				SenderID:            v.savedNote.SenderID,
				ReceiverID:          v.savedNote.ReceiverID,
				SenderMerkleProof:   v.savedNote.SenderMerkleProof,
				ReceiverMerkleProof: nil,
				Calluuid:            v.savedNote.Calluuid,
			},
		}
	}

	return sresponse
}
