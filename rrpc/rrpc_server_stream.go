package rrpc

import (
	"encoding/binary"
	"io"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

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

	if err := s.relaystreams.addAt(relayIndex, stream); err != nil {
		return err
	}

	// used by the server to send messages to the relay.
	go func() {
		for tosend := range s.relaystreams.getChan(relayIndex) {
			if err := stream.Send(tosend); err != nil {
				s.relaystreams.removeAt(relayIndex)
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
			s.relaystreams.sendTo(relayIndex,
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
	if err := s.verifier.Verify(parcel.Note.SenderID, parcel); err != nil {
		return err
	}
	(*senderNote)(parcel.Note).pushCert(tmpCert)

	// verify the note contains a valid signature too. Otherwise, you cannot fill it!
	return s.verifier.Verify(parcel.Note.ReceiverID, (*senderNote)(parcel.Note))
}

type parcelCollection struct {
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

	tasks := map[string]*parcelCollection{}

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
				s.relaystreams.sendTo(int(task.RelayIndex), &CallStreamResponse{
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
				response := s.runTask(v)
				_ = response

				//for _, streamTask := range response {
				//	s.signAndStreamTasks <- streamTask
				//}
			}()

		}
	}
}

func (s *Server) getOrCreateTask(tasks map[string]*parcelCollection, parcel *Parcel) (*parcelCollection, error) {
	service, methodDesc, err := s.getServiceAndMethodDesc(parcel.Method)
	if err != nil {
		return nil, err
	}

	v, ok := tasks[parcel.Note.Calluuid]
	if !ok {
		v = &parcelCollection{
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

func (s *Server) runTask(v *parcelCollection) []*CallStreamResponse {
	payload, err := s.reconstructParcels(v)
	if err != nil {
		return nil // s.prepareErrorResponse(j, err)
	}

	_ = payload
	//
	//handlerResponse, err := j.call.methodHandler(j.call.service.server, ctx, createDecodeFunc(payload))
	//if err != nil {
	//	return s.prepareErrorResponse(j, err)
	//}
	//
	//responses, err := s.prepareCallResponse(handlerResponse, j)
	//if err != nil {
	//	return s.prepareErrorResponse(j, err)
	//}

	return nil
}

func (s *Server) reconstructParcels(v *parcelCollection) (interface{}, error) {
	if len(v.parcels) == 0 {
		return nil, status.Error(codes.Internal, "reached reconstruction with 0 parcels")
	}

	msgSize := binary.LittleEndian.Uint32(v.parcels[0].MessageLength)

	shards := s.decoderEncoder.NewShards()

	for i := 0; i < len(v.parcels); i++ {
		(*eccClientParcel)(v.parcels[i]).PutIntoShards(shards)

		v.parcels[i] = nil
	}

	data, err := s.decoderEncoder.AuthReconstruct(shards, int(msgSize))
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "server reconstruction failure: %v", err)
	}

	return data, nil
}
