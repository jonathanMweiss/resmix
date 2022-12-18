package rrpc

import (
	"context"
	"fmt"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RelayConn struct {
	*grpc.ClientConn
	RelayClient
	context.Context
	context.CancelFunc
	sendProofChan chan *Proof
	sync.WaitGroup
	sync.Map
}

func NewRelayConn(address string) (*RelayConn, error) {
	cc, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	relayClient := NewRelayClient(cc)

	ctx, cancel := context.WithCancel(context.Background())
	r := &RelayConn{
		ClientConn:  cc,
		RelayClient: relayClient,
		Context:     ctx,
		CancelFunc:  cancel,
	}

	sendProofStream, err := relayClient.SendProof(ctx)
	if err != nil {
		return nil, err
	}

	r.WaitGroup.Add(1)
	go func() {
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
	}()
	// setup streams. then you can send through functions on this client, and expect a response when ready.

	return r, nil
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
