package rrpc

import (
	"context"
	"fmt"
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

	waitingTasks sync.Map // [uuid, chan Response of type?]

	wg sync.WaitGroup
	//verifier *MerkleCertVerifier

	identifier []byte
	serverID   []byte

	directCallSendChannel chan *requestWithResponse[*Request]

	context.Context
	context.CancelFunc
}
type requestWithResponse[T interface{}] struct {
	In        T
	Err       chan error
	StartTime time.Time
}

func (c *client) Close() error {
	c.CancelFunc()
	return nil
}

func (c *client) setServerStream() error {
	stream, err := c.serverClient.DirectCall(AddIPToContext(context.Background(), c.myAddr))
	if err != nil {
		return err
	}

	c.wg.Add(3)
	go func() {
		defer c.wg.Done()
		for {
			msg, err := stream.Recv()
			if err != nil {
				fmt.Println("client::streamSend error: ", err.Error())
				return
			}
			if err := c.VerifyAndDispatch(msg); err != nil {
				fmt.Println("client::streamSend:: dispatch error: ", err.Error())
				return
			}

			c.network.PublishProof(&Proof{
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

	go garbageCollectSyncMap[*requestWithResponse[*Request]](c.Context, &c.wg, &c.waitingTasks)

	return nil
}

// VerifyAndDispatch will verify the response, any critical error will result in closing of the stream
func (c *client) VerifyAndDispatch(msg *DirectCallResponse) error {
	v, ok := c.waitingTasks.LoadAndDelete(msg.Note.Calluuid)
	if !ok {
		fmt.Println("client::VerifyAndDispatch: could not find task!")
		return nil
	}

	reqst, ok := v.(*requestWithResponse[*Request])
	if !ok {
		panic("client::VerifyAndDispatch: could not cast task!")
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
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	c := client{
		myAddr:                ownAddress,
		serverAddr:            serverAddress,
		serverClient:          NewServerClient(cc),
		network:               network,
		secretKey:             key,
		encoderDecoder:        encoderDecoder,
		verifier:              NewVerifier(1),
		waitingTasks:          sync.Map{},
		wg:                    sync.WaitGroup{},
		identifier:            key.Public(),
		serverID:              serverPk,
		directCallSendChannel: make(chan *requestWithResponse[*Request], 10),
		Context:               ctx,
		CancelFunc:            cancel,
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

	reqst := &requestWithResponse[*Request]{
		In:        req,
		Err:       make(chan error),
		StartTime: time.Now(),
	}

	c.waitingTasks.Store(req.Uuid, reqst)
	c.directCallSendChannel <- reqst

	return <-reqst.Err
}
