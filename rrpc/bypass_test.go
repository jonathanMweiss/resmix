package rrpc

import (
	"context"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"net"
	"testing"
)

type grpcTestInterface interface {
	Check() *Parcel
}

type grpcSimpleService struct {
	message string
}

func (g grpcSimpleService) Check() *Parcel {
	return &Parcel{Payload: [][]byte{[]byte(g.message)}}
}

func TestGrpcServiceWorks(t *testing.T) {

	service := &grpcSimpleService{
		message: "testService",
	}

	services := Services{
		"testService": {
			serverType: (*grpcTestInterface)(nil),
			server:     service,
			methodDescriptors: map[string]*MethodDesc{
				"testMethod": {
					Name: "testMethod",
					Handler: func(server interface{}, ctx context.Context, dec func(interface{}) error) (interface{}, error) {

						return server.(grpcTestInterface).Check(), nil
					},
				},
			},
		},
	}

	addr := "localhost:6060"
	l, err := net.Listen("tcp", addr)
	require.NoError(t, err)

	srvr := newGrpcServer(services)
	go srvr.Serve(l)

	defer srvr.Stop()

	c, err := newGrpcClient(addr, grpc.WithInsecure())
	require.NoError(t, err)

	defer c.Close()

	req := &Request{
		Args:    nil,
		Reply:   new(Parcel),
		Method:  "testService/testMethod",
		Uuid:    "1234",
		Context: context.Background(),
	}

	chn, _ := c.AsyncDirectCall(req)

	e := <-chn
	require.NoError(t, e)

	msg := string(req.Reply.(*Parcel).Payload[0])
	require.Equal(t, service.message, msg)
}
