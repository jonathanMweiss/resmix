package rrpc

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/jonathanMweiss/resmix/internal/crypto"
	"github.com/stretchr/testify/require"
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

type clientTestSetup struct {
	network    Network
	sk         crypto.PrivateKey
	serverAddr string
	srvrs      []RrpcServer
}

func (c *clientTestSetup) start(t *testing.T) {
	require.NoError(t, c.network.RelayDial())
}

func (c *clientTestSetup) releaseResources() {
	for _, srvr := range c.srvrs {
		srvr.Stop()
	}
}

func newClientTestSetup(t *testing.T) clientTestSetup {
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

	network := NewNetwork(netdata, sk)

	srvrs := make([]RrpcServer, 0, len(netdata.Servers()))

	for _, s := range netdata.Servers() {
		l, err := net.Listen("tcp", s)
		require.NoError(t, err)

		srvr := NewServerService(sk, srvc, network)
		srvrs = append(srvrs, srvr)

		go func() {
			require.NoError(t, srvr.Serve(l))
		}()
	}

	return clientTestSetup{
		network:    network,
		sk:         sk,
		serverAddr: serverAddr,
		srvrs:      srvrs,
	}
}

func TestDirectCall(t *testing.T) {
	setup := newClientTestSetup(t)

	setup.start(t)
	defer setup.releaseResources()

	// Ensuring the network dials to all relays.
	c := NewClient(setup.sk, setup.serverAddr, setup.network)
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
