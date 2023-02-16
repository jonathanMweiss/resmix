package rrpc

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"github.com/jonathanMweiss/resmix/internal/msync"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/credentials/insecure"
	"math"
	"sync"
	"time"

	"github.com/jonathanMweiss/resmix/internal/crypto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// client is responsible for connecting to its main server,
// to connect through relays, it expects to receive some object that can communicate via relays.
type client struct {
	myAddr       string
	serverAddr   string
	serverClient ServerClient
	network      Coordinator
	secretKey    crypto.PrivateKey

	waitingTasks msync.Map[string, *rqstWithErr] // [uuid, chan Response of type?]

	wg         sync.WaitGroup
	bufferpool sync.Pool

	identifier []byte
	serverID   []byte

	directCallSendChannel chan *rqstWithErr

	context.Context
	context.CancelFunc
	log *logrus.Entry
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
	return nil
}

func (c *client) setServerStream() error {
	stream, err := c.serverClient.DirectCall(addIPToContext(c.Context, c.myAddr))
	if err != nil {
		return err
	}

	c.wg.Add(3)
	go func() {
		defer c.wg.Done()

		entry := c.log.WithField("method", "client::streamRecv")

		for {
			msg, err := stream.Recv()
			if isEOFFromServer(err) {
				entry.Debugln("client::streamSend closing")
				return
			}
			if err != nil {
				entry.Errorln("client::streamSend error: ", err.Error())
				return
			}
			if err := c.VerifyAndDispatch(msg); err != nil {
				entry.Warnln("client::streamSend:: dispatch error: ", err.Error())
				continue
			}

			c.network.getRelayGroup().PublishProof(&Proof{
				ServerHostname:   c.serverAddr,
				WorkExchangeNote: msg.Note,
			})
		}
	}()

	go func() {
		defer c.wg.Done()

		entry := c.log.WithField("method", "client::streamSend")

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
					entry.Errorln("signing error: ", err.Error())
					task.Err <- err
					return
				}

				if err := stream.Send(rqst); err != nil {
					// This is a send task, no response just yet, because we should wait on response too
					task.Err <- err
					entry.Errorln("client::streamSend:: send error: ", err.Error())
					continue
				}
			case <-c.Done():
				entry.Debugln("client::streamSend:: closing direct call send stream")
				return
			}

		}
	}()

	go func() {
		defer c.wg.Done()

		foreverCleanup(c.Context, &c.waitingTasks)
	}()

	return nil
}

// VerifyAndDispatch will verify the response, any critical error will result in closing of the stream
func (c *client) VerifyAndDispatch(msg *DirectCallResponse) error {
	reqst, ok := c.waitingTasks.LoadAndDelete(msg.Note.Calluuid)
	if !ok {
		c.log.Debugln("client::VerifyAndDispatch: could not find task!")
		return nil
	}

	var err error
	defer func() {
		reqst.Err <- err
		close(reqst.Err)
	}()

	err = c.network.getVerifier().Verify(c.serverID, (*receiverNote)(msg.Note))
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

func newClient(serverAddress string, network Coordinator) *client {
	ownAddress := network.GetHostname(network.GetSecretKey().Public())
	serverPk, err := network.GetPublicKey(serverAddress)
	if err != nil {
		panic(err)
	}

	cc, err := grpc.Dial(serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	c := client{
		log:          logrus.WithFields(logrus.Fields{"component": "rrpc.client", "from": ownAddress, "to": serverAddress}),
		myAddr:       ownAddress,
		serverAddr:   serverAddress,
		serverClient: newServerClient(cc),
		network:      network,
		secretKey:    network.GetSecretKey(),

		waitingTasks:          msync.Map[string, *rqstWithErr]{},
		wg:                    sync.WaitGroup{},
		identifier:            network.GetSecretKey().Public(),
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
	chn, err := c.AsyncDirectCall(req)
	if err != nil {
		return err
	}

	return <-chn
}

func (c *client) AsyncDirectCall(req *Request) (<-chan error, error) {
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

	out, err := c.network.getRelayGroup().RobustRequest(req, robustCallRequests)
	if err != nil {
		return err
	}

	payload, err := c.reconstruct(out)
	if err != nil {
		return err
	}

	req.marshaledReply = payload
	return req.unpack()
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

	chunks, err := c.network.getErrorCorrectionCode().AuthEncode(margs)
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

func (c *client) reconstruct(parcels []*CallStreamResponse) ([]byte, error) {
	//// safety measure.
	if len(parcels) <= 0 {
		return nil, status.Error(codes.DataLoss, "no parcels received")
	}

	if len(parcels) == 0 {
		panic("no parcels received")
	}
	if parcels[0] == nil {
		panic("What?!")
	}
	msgSize := binary.LittleEndian.Uint32(parcels[0].Response.MessageLength)

	shards := c.network.getErrorCorrectionCode().NewShards()

	for _, parcel := range parcels {
		(*eccServerParcel)(parcel.Response).PutIntoShards(shards)
	}

	data, err := c.network.getErrorCorrectionCode().AuthReconstruct(shards, int(msgSize))
	if err != nil {
		return nil, status.Error(codes.Internal, "client: reconstructing failed: "+err.Error())
	}

	return data, nil
}
