package rrpc

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/jonathanMweiss/resmix/internal/crypto"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

const _serverport = "5005"
const _minimal_service_reply = "minimal service reply"

var srvc Services = Services{
	"testService": {
		server: (new)(bool),
		methodDescriptors: map[string]*MethodDesc{
			"testMethod": {
				Name: "testMethod",
				Handler: func(server interface{}, ctx context.Context, dec func(interface{}) error) (interface{}, error) {
					return _minimal_service_reply, nil
				},
			},
		},
	},
}

func TestName(t *testing.T) {
	almostAddr := "localhost:" + _serverport
	serverAddr := "localhost:" + _serverport + "1"

	sk, pk, err := crypto.GenerateKeys()
	require.NoError(t, err)
	netdata := NewNetData(&NetworkConfig{
		Tau: 2,
		ServerConfigs: []ServerData{
			{
				Address:   almostAddr + "1",
				Publickey: pk,
			},
			{
				Address:   almostAddr + "2",
				Publickey: pk,
			},
			{
				Address:   almostAddr + "3",
				Publickey: pk,
			},
		},
	})
	for _, s := range netdata.Servers() {
		srvr := NewServerService(sk, srvc, netdata)

		l, err := net.Listen("tcp", s)
		require.NoError(t, err)

		gsrvr := grpc.NewServer()
		RegisterServerServer(gsrvr, srvr)
		wg := &sync.WaitGroup{}
		wg.Add(1)
		defer wg.Wait()

		defer gsrvr.Stop()
		go func() {
			defer wg.Done()
			require.NoError(t, gsrvr.Serve(l))
		}()
	}

	network, err := NewNetwork(netdata, sk)
	require.NoError(t, err)
	c := NewClient(sk, serverAddr, network)

	req := &Request{
		Args:    nil,
		Reply:   new(string),
		Method:  "testService/testMethod",
		Uuid:    "1234",
		Context: context.Background(),
	}

	require.NoError(t, c.DirectCall(req))
	require.Equal(t, _minimal_service_reply, *(req.Reply.(*string)))
	time.Sleep(time.Second)
}
