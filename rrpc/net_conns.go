package rrpc

import (
	"context"
	"github.com/jonathanMweiss/resmix/internal/msync"
	"github.com/sirupsen/logrus"
	"io"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type RelayConn struct {
	log *logrus.Entry

	*grpc.ClientConn
	RelayClient
	context.Context
	context.CancelFunc
	sync.WaitGroup

	sendProofChan chan *Proof
	requests      chan *RelayStreamRequest
	liveTasks     msync.Map[string, relayConnRequest]
	index         int
}

type relayResponse *RelayStreamResponse

type relayConnRequest struct {
	*RelayRequest
	response chan relayResponse
	time.Time
}

func (r relayConnRequest) GetStartTime() time.Time {
	return r.Time
}

func (rqst relayConnRequest) PrepareForDeletion() {
	rqst.response <- &RelayStreamResponse{
		RelayStreamError: status.New(codes.Canceled, "response timed out").Proto(),
		Uuid:             rqst.RelayRequest.Parcel.Note.Calluuid,
	}
}

func newRelayConn(ctx context.Context, address string, index int) (*RelayConn, error) {
	cc, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	relayClient := newRelayClient(cc)

	ctx, cancel := context.WithCancel(ctx)
	r := &RelayConn{
		log: logrus.WithFields(logrus.Fields{"component": "rrpc.relayConn", "index": index}),

		ClientConn:  cc,
		RelayClient: relayClient,
		Context:     ctx,
		CancelFunc:  cancel,
		WaitGroup:   sync.WaitGroup{},

		index: index,

		sendProofChan: make(chan *Proof, 100),
		requests:      make(chan *RelayStreamRequest, 100),
		liveTasks:     msync.Map[string, relayConnRequest]{},
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
		defer r.WaitGroup.Done()

		foreverCleanup(r.Context, &r.liveTasks)
	}()

	return r, nil
}

func (r *RelayConn) proofSendingStream(sendProofStream Relay_SendProofClient) {
	defer r.WaitGroup.Done()

	for {
		select {
		case <-r.Context.Done():
			if err := sendProofStream.CloseSend(); err != nil {
				r.log.Warnln("closing streamProof failed:", err)
			}

			return
		case prf := <-r.sendProofChan:
			if err := sendProofStream.Send(&Proofs{
				Proofs: []*Proof{prf},
			}); err != nil {
				r.log.Warnln("sending proof error:", err)
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
	r.liveTasks.Store(rqst.Parcel.Note.Calluuid, rqst)
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

	entry := r.log.WithField("method", "parcelStream")

	var rqst *RelayStreamRequest
	for {
		select {
		case <-r.Context.Done():
			if err := stream.CloseSend(); err != nil {
				entry.Errorln("closing stream failed: ", err)
			}

			return
		case rqst = <-r.requests:
			err := stream.Send(rqst)
			if err == io.EOF {
				entry.Debugln("stream closing")

				return
			}

			if err != nil {
				entry.Errorln("stream error: ", err)
			}
		}
	}
}

func (r *RelayConn) receiveParcels(stream Relay_RelayStreamClient) {
	defer r.WaitGroup.Done()

	entry := r.log.WithField("method", "receiveParcels")

	for {

		out, err := stream.Recv()
		if isEOFFromServer(err) {
			return
		}

		if err != nil {
			entry.Errorln("stream error: ", err)

			return
		}

		task, ok := r.liveTasks.LoadAndDelete(out.Uuid)
		if !ok {
			continue
		}

		if task.response == nil {
			continue
		}

		task.response <- out
	}
}

func isEOFFromServer(err error) bool {
	if err == io.EOF {
		return true
	}

	if err == io.ErrUnexpectedEOF {
		return true
	}

	st, ok := status.FromError(err)
	if !ok {
		return false

	}

	return st.Code() == codes.Unavailable || st.Code() == codes.Canceled
}

type ServerConn struct {
	clientConn ServerClient
	context    context.Context
	cancel     context.CancelFunc
	cc         *grpc.ClientConn
	stream     Server_CallStreamClient

	toSend     chan *CallStreamRequest
	outputChan chan *CallStreamResponse
	wg         sync.WaitGroup
	log        *logrus.Entry
}

func (c *ServerConn) Close() error {
	c.cancel() // closing any running streams!
	return c.cc.Close()
}

func (c *ServerConn) send(msg *CallStreamRequest) {
	select {
	case c.toSend <- msg:
	case <-c.context.Done():
	}
}

func newServerConn(ctx context.Context, address string, output chan *CallStreamResponse) (*ServerConn, error) {
	cc, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	con := newServerClient(cc)

	connctx, cancel := context.WithCancel(ctx)

	stream, err := con.CallStream(connctx)
	if err != nil {
		cancel()

		return nil, cc.Close()
	}

	conn := &ServerConn{
		log:        logrus.WithFields(logrus.Fields{"component": "rrpc.serverconn"}),
		cc:         cc,
		clientConn: con,
		stream:     stream,
		context:    connctx,
		cancel:     cancel,

		wg:         sync.WaitGroup{},
		toSend:     make(chan *CallStreamRequest, 100),
		outputChan: output,
	}

	conn.wg.Add(2)

	go func() {
		defer conn.wg.Done()

		var tosend *CallStreamRequest

		entry := conn.log.WithField("method", "StreamSend")

		for {
			select {
			case <-connctx.Done():
				if err := conn.stream.CloseSend(); err != nil {
					entry.Errorln("closing stream failed: ", err)
				}

				return

			case tosend = <-conn.toSend:
			}

			err := conn.stream.Send(tosend)
			if isEOFFromServer(err) {
				return
			}

			if err != nil {
				conn.log.Errorln("stream error:", err)
			}
		}
	}()

	go func() {
		defer conn.wg.Done()

		entry := conn.log.WithField("method", "StreamRecv")

		for {
			o, err := conn.stream.Recv()
			if isEOFFromServer(err) {
				return
			}

			if err != nil {
				entry.Errorln("stream error:", err)

				continue
			}

			conn.outputChan <- o
		}
	}()

	return conn, nil
}
