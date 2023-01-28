package rrpc

import (
	"io"
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
			s.streamErrorBack(request, err)
			continue
		}
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

func (s *Server) streamErrorBack(request *CallStreamRequest, err error) {

}
