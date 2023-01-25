package rrpc

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type RelayConn struct {
	*grpc.ClientConn
	RelayClient
	context.Context
	context.CancelFunc
	sync.WaitGroup

	sendProofChan chan *Proof
	requests      chan *RelayStreamRequest
	liveTasks     sync.Map
	index         int
}

type relayResponse *RelayStreamResponse

type relayConnRequest struct {
	*RelayRequest
	response chan relayResponse
	time.Time
}

func NewRelayConn(address string, index int) (*RelayConn, error) {
	cc, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	relayClient := NewRelayClient(cc)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	r := &RelayConn{
		ClientConn:  cc,
		RelayClient: relayClient,
		Context:     ctx,
		CancelFunc:  cancel,
		WaitGroup:   sync.WaitGroup{},

		index: index,

		sendProofChan: make(chan *Proof, 100),
		requests:      make(chan *RelayStreamRequest, 100),
		liveTasks:     sync.Map{},
	}

	sendProofStream, err := relayClient.SendProof(ctx)
	if err != nil {
		return nil, err
	}

	relayStreamClient, err := relayClient.RelayStream(ctx)
	if err != nil {
		return nil, err
	}

	r.WaitGroup.Add(4)

	go r.proofSendingStream(sendProofStream)

	go r.parcelStream(relayStreamClient)

	go r.receiveParcels(relayStreamClient)

	go func() {
		// todo find way to reuse code here...
		cleanTime := time.NewTicker(time.Second * 5)

		for {
			select {
			case <-r.Context.Done():
				return
			case <-cleanTime.C:
				cur := time.Now()
				r.liveTasks.Range(func(key, value interface{}) bool {
					if cur.Sub(value.(relayConnRequest).Time) > time.Second*5 {
						r.liveTasks.Delete(key)
					}

					return true
				})
			}
		}
	}()

	return r, nil
}

func (r *RelayConn) proofSendingStream(sendProofStream Relay_SendProofClient) {
	defer r.WaitGroup.Done()

	for {
		select {
		case <-r.Context.Done():
			if err := sendProofStream.CloseSend(); err != nil {
				fmt.Println("closing streamProof failed:", err)
			}

			return
		case prf := <-r.sendProofChan:
			if err := sendProofStream.Send(&Proofs{
				Proofs: []*Proof{prf}, // todo send more than one.
			}); err != nil {
				fmt.Println("sending proof error:", err)
			}
		}
	}
}

func (r *RelayConn) SendProof(proof *Proof) {
	select {
	case r.sendProofChan <- proof:
		// RelayConn is closed:
	case <-r.Context.Done():
		// if the sendProofChan is blocked -> don't wait on it...
		// better to avoid it altogether.
	default:
	}
}

func (r *RelayConn) Close() error {
	r.CancelFunc()
	r.WaitGroup.Wait()

	return r.ClientConn.Close()
}

func (r *RelayConn) cancelRequest(uuid string) {
	r.liveTasks.Delete(uuid)
}

func (r *RelayConn) sendRequest(rqst relayConnRequest) {
	r.liveTasks.Store(rqst.Parcel.Note.Calluuid, rqst.response)
	select {
	case <-r.Context.Done():
		return
	case r.requests <- &RelayStreamRequest{
		Request: rqst.RelayRequest,
	}:
	}
}

func (r *RelayConn) parcelStream(stream Relay_RelayStreamClient) {
	defer r.WaitGroup.Done()

	var rqst *RelayStreamRequest
	for {
		select {
		case <-r.Context.Done():
			return
		case rqst = <-r.requests:
			err := stream.Send(rqst)
			if err == io.EOF {
				fmt.Printf("relay(%d) parcel stream closeing\n", r.index)
				return
			}
			if err != nil {
				fmt.Printf("relay(%d) parcel stream error: %v\n", r.index, err)
			}
		}
	}
}

func (r *RelayConn) receiveParcels(stream Relay_RelayStreamClient) {
	defer r.WaitGroup.Done()

	for {

		select {
		case <-r.Context.Done():
			return
		default:
		}

		out, err := stream.Recv()
		if isEOFFromServer(err) {
			return
		}

		if err != nil {
			fmt.Printf("relay(%d) receive parcel stream error: %v\n", r.index, err)
			continue
		}

		task, ok := r.liveTasks.LoadAndDelete(out.Uuid)
		if !ok {
			continue
		}

		if task.(relayConnRequest).response == nil {
			continue
		}

		task.(relayConnRequest).response <- out
	}
}

func isEOFFromServer(err error) bool {
	if err == io.EOF {
		return true
	}

	if err == io.ErrUnexpectedEOF {
		return true
	}
	if st, ok := status.FromError(err); ok {
		return st.Code() == codes.Unavailable
	}
	return false
}
