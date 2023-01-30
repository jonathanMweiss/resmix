package rrpc

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"github.com/jonathanMweiss/resmix/internal/msync"
	"math"
	"sync"
	"time"

	"github.com/jonathanMweiss/resmix/internal/crypto"
	"github.com/jonathanMweiss/resmix/internal/ecc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// client is responsible for connecting to its main server,
// to connect through relays, it expects to receive some object that can communicate via relays.
type client struct {
	myAddr         string
	serverAddr     string
	serverClient   ServerClient
	network        Network
	secretKey      crypto.PrivateKey
	encoderDecoder ecc.VerifyingEncoderDecoder
	verifier       *MerkleCertVerifier

	waitingTasks msync.Map[string, *rqstWithErr] // [uuid, chan Response of type?]

	wg         sync.WaitGroup
	bufferpool sync.Pool

	identifier []byte
	serverID   []byte

	directCallSendChannel chan *rqstWithErr

	context.Context
	context.CancelFunc
}

// implementing GCable interface
type rqstWithErr struct {
	In        *Request
	Err       chan error
	StartTime time.Time
}

func (r *rqstWithErr) GetStartTime() time.Time {
	return r.StartTime
}

func (rqst *rqstWithErr) PrepareForDeletion() {
	rqst.Err <- status.Error(codes.Canceled, "response timed out")
}

func (c *client) Close() error {
	c.CancelFunc()
	close(c.verifier.done)
	return nil
}

func (c *client) setServerStream() error {
	stream, err := c.serverClient.DirectCall(AddIPToContext(c.Context, c.myAddr))
	if err != nil {
		return err
	}

	c.wg.Add(3)
	go func() {
		defer c.wg.Done()
		for {
			msg, err := stream.Recv()
			if isEOFFromServer(err) {
				fmt.Println("client::streamSend closing")
				return
			}
			if err != nil {
				fmt.Println("client::streamSend error: ", err.Error())
				return
			}
			if err := c.VerifyAndDispatch(msg); err != nil {
				fmt.Println("client::streamSend:: dispatch error: ", err.Error())
				continue
			}

			c.network.GetRelayGroup().PublishProof(&Proof{
				ServerHostname:   c.serverAddr,
				WorkExchangeNote: msg.Note,
			})
		}
	}()

	go func() {
		defer c.wg.Done()
		for {
			select {
			case task := <-c.directCallSendChannel:
				rqst := &DirectCallRequest{
					Method:  task.In.Method,
					Payload: task.In.marshaledArgs,
					Note: &ExchangeNote{
						SenderID:            c.identifier,
						ReceiverID:          c.serverID,
						SenderMerkleProof:   nil,
						ReceiverMerkleProof: nil,
						Calluuid:            task.In.Uuid,
					},
				}
				if err := merkleSign([]MerkleCertifiable{(*senderNote)(rqst.Note)}, c.secretKey); err != nil {
					fmt.Println("streaming error: ", err.Error())
					task.Err <- err
					return
				}

				if err := stream.Send(rqst); err != nil {
					// This is a send task, no response just yet, because we should wait on response too
					task.Err <- err
					fmt.Println("streaming error: ", err.Error())
					continue
				}
			case <-c.Done():
				fmt.Println("closing direct call send stream")
				return
			}

		}
	}()

	go func() {
		defer c.wg.Done()

		const ttl = time.Second * 5

		dropFromMap := time.NewTicker(ttl)
		for {
			select {
			case <-dropFromMap.C:
				cleanmapAccordingToTTL(&c.waitingTasks, ttl)
			case <-c.Context.Done():
				return
			}
		}
	}()

	return nil
}

// VerifyAndDispatch will verify the response, any critical error will result in closing of the stream
func (c *client) VerifyAndDispatch(msg *DirectCallResponse) error {
	reqst, ok := c.waitingTasks.LoadAndDelete(msg.Note.Calluuid)
	if !ok {
		fmt.Println("client::VerifyAndDispatch: could not find task!")
		return nil
	}

	var err error
	defer func() {
		reqst.Err <- err
		close(reqst.Err)
	}()

	err = c.verifier.Verify(c.serverID, (*receiverNote)(msg.Note))
	if err != nil {
		return err
	}

	switch res := msg.Result.(type) {
	case *DirectCallResponse_RpcError:
		err = status.Error(codes.Code(res.RpcError.GetCode()), res.RpcError.GetMessage())
	case *DirectCallResponse_Payload:
		reqst.In.marshaledReply = res.Payload
		err = reqst.In.unpack()
	}

	return nil
}

func NewClient(key crypto.PrivateKey, serverAddress string, network Network) *client {
	ownAddress := network.GetHostname(key.Public())
	serverPk, err := network.GetPublicKey(serverAddress)
	if err != nil {
		panic(err)
	}

	cc, err := grpc.Dial(serverAddress, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	encoderDecoder, err := network.NewErrorCorrectionCode()
	//if err != nil {
	//	panic(err)
	//}

	ctx, cancel := context.WithCancel(context.Background())
	c := client{
		myAddr:                ownAddress,
		serverAddr:            serverAddress,
		serverClient:          NewServerClient(cc),
		network:               network,
		secretKey:             key,
		encoderDecoder:        encoderDecoder,
		verifier:              NewVerifier(1),
		waitingTasks:          msync.Map[string, *rqstWithErr]{},
		wg:                    sync.WaitGroup{},
		identifier:            key.Public(),
		serverID:              serverPk,
		directCallSendChannel: make(chan *rqstWithErr, 10),
		Context:               ctx,
		CancelFunc:            cancel,

		bufferpool: sync.Pool{New: func() interface{} { return new(bytes.Buffer) }},
	}

	if err := c.setServerStream(); err != nil {
		panic(err)
	}
	return &c
}

func (c *client) DirectCall(req *Request) error {
	if err := req.pack(); err != nil {
		return err
	}

	reqst := &rqstWithErr{
		In:        req,
		Err:       make(chan error),
		StartTime: time.Now(),
	}

	c.waitingTasks.Store(req.Uuid, reqst)
	c.directCallSendChannel <- reqst

	return <-reqst.Err
}

func (c *client) AsyncDirectCall(req *Request) (chan<- error, error) {
	if err := req.pack(); err != nil {
		return nil, err
	}

	reqst := &rqstWithErr{
		In:        req,
		Err:       make(chan error),
		StartTime: time.Now(),
	}

	c.waitingTasks.Store(req.Uuid, reqst)
	c.directCallSendChannel <- reqst

	return reqst.Err, nil
}

func (c *client) RobustCall(req *Request) error {
	if err := req.pack(); err != nil {
		return err
	}

	robustCallRequests, err := c.requestIntoRobust(req)
	if err != nil {
		return err
	}

	// sign them
	signables := make([]MerkleCertifiable, 0, 2*len(robustCallRequests))
	for i := range robustCallRequests {
		signables = append(signables, robustCallRequests[i])
	}
	for i := range robustCallRequests {
		signables = append(signables, (*senderNote)(robustCallRequests[i].Parcel.Note))
	}

	if err := merkleSign(signables, c.secretKey); err != nil {
		return err
	}

	out, err := c.network.GetRelayGroup().RobustRequest(req, robustCallRequests)
	if err != nil {
		return err
	}

	// TODO: reconstruct!
	c.reconstruct(out)
	return nil
}

func (c *client) requestIntoRobust(rq *Request) ([]*RelayRequest, error) {

	if err := rq.pack(); err != nil {
		return nil, err
	}

	margs := rq.marshaledArgs

	if len(margs) > math.MaxUint32 {
		return nil, fmt.Errorf("message size: %v is too big", len(margs))
	}

	msgLength := make([]byte, 4)
	binary.LittleEndian.PutUint32(msgLength, uint32(len(margs)))

	chunks, err := c.encoderDecoder.AuthEncode(margs)
	if err != nil {
		return nil, fmt.Errorf("failure in encoding, client side: %v", err)
	}

	relayRequests := make([]*RelayRequest, len(c.network.Servers()))
	for i := range c.network.Servers() {
		relayRequests[i] = &RelayRequest{
			Parcel: &Parcel{
				Method:     rq.Method,
				RelayIndex: int32(i),

				Note: &ExchangeNote{
					SenderID:            c.identifier,
					ReceiverID:          c.serverID,
					SenderMerkleProof:   nil, // will be signed soon
					ReceiverMerkleProof: nil,
					Calluuid:            rq.Uuid,
				},
				MessageLength: msgLength,
			},
		}
		// O(1) work.
		(*eccClientParcel)(relayRequests[i].Parcel).InsertECCPayload(chunks, i)
	}

	return relayRequests, nil
}

func (c *client) reconstruct(out []*CallStreamResponse) {
	// TODO
	panic("implement me")
}
