package rrpc

import (
	"fmt"
	"io"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CallStream is the part in the server that handles incoming rRPC parcels, forwards it to the server's collector to handle.
func (s *Server) CallStream(stream Server_CallStreamServer) error {
	//TODO implement me
	for {
		request, err := stream.Recv()
		if err == io.EOF {
			return nil
		}

		if err != nil {
			return err
		}

		if err := s.validateParcel(request.Parcel); err != nil {
			s.streamBackResponse(&CallStreamResponse{
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

func (s *Server) validateParcel(parcel *Parcel) error {
	tmpCert := (*senderNote)(parcel.Note).popCert()
	if err := s.verifier.Verify(parcel.Note.SenderID, parcel); err != nil {
		return err
	}
	(*senderNote)(parcel.Note).pushCert(tmpCert)

	// verify the note contains a valid signature too. Otherwise, you cannot fill it!
	return s.verifier.Verify(parcel.Note.ReceiverID, (*senderNote)(parcel.Note))
}

func (s *Server) streamBackResponse(*CallStreamResponse) {

}

type parcelCollection struct {
	parcels   []*Parcel
	savedNote *ExchangeNote
	startTime time.Time
}

// collector reduces the incoming parcels into a task.
// once we reach enough parcels, we can decode them and run them.
func (s *Server) collector() {
	defer s.WaitGroup.Done()

	ttl := time.Second * 5
	timeToLiveTicker := time.NewTicker(ttl)

	tasks := map[string]parcelCollection{}
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
			fmt.Println("Collected")
			_ = task
			//j, err := s.getOrCreateJob(task.newRequest.Parcel)
			//if err != nil {
			//	s.sendError(status.Newf(codes.Internal, err.Error()), task.newRequest, task.sendingRelay)
			//	continue
			//}
			//
			//j.call.parcels = append(j.call.parcels, task.newRequest.Parcel)
			//if len(j.call.parcels) < s.seminet.MinimalRelayedParcels() {
			//	continue
			//}
			//
			//delete(s.jobs, j.callId)
			//
			//// run the job on a different goroutine.
			//go func() {
			//	response := s.runClientJob(j)
			//	for _, streamTask := range response {
			//		s.signAndStreamTasks <- streamTask
			//	}
			//	j.call.parcels = nil
			//}()
		}
	}
}
