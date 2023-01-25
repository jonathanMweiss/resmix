package rrpc

import (
	"context"
	"net"
	"strconv"
	"testing"

	"github.com/jonathanMweiss/resmix/internal/crypto"
	"github.com/stretchr/testify/require"
)

const _serverport = "600"
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
	require.NoError(t, c.network.Dial())
}

func (c *clientTestSetup) releaseResources() {
	for _, srvr := range c.srvrs {
		srvr.Stop()
	}
}

func newClientTestSetup(t *testing.T) clientTestSetup {
	almostAddr := "localhost:" + _serverport
	serverAddr := "localhost:" + _serverport + "10"

	netconf := &NetworkConfig{
		Tau:           2,
		ServerConfigs: []ServerData{},
	}
	sks := []crypto.PrivateKey{}
	for i := 1; i <= 3; i++ {
		sk, pk, err := crypto.GenerateKeys()
		require.NoError(t, err)

		sks = append(sks, sk)
		netconf.ServerConfigs = append(netconf.ServerConfigs, ServerData{
			Address:   almostAddr + strconv.Itoa(i*10),
			Publickey: pk,
		})
	}
	netdata := NewNetData(netconf)

	network := NewNetwork(netdata, sks[0])

	srvrs := make([]RrpcServer, 0, len(netdata.Servers()))

	for i, s := range netdata.Servers() {
		l, err := net.Listen("tcp", s)
		require.NoError(t, err)

		srvr := NewServerService(sks[i], srvc, network)
		srvrs = append(srvrs, srvr)

		go func() {
			if srvr.Serve(l) != nil {
				panic("Server failed to serve")
			}
		}()
	}

	return clientTestSetup{
		network:    network,
		sk:         sks[0],
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
	defer c.Close()
	for i := 0; i < 10; i++ {
		req := &Request{
			Args:    nil,
			Reply:   new(string),
			Method:  "testService/testMethod",
			Uuid:    "1234",
			Context: context.Background(),
		}
		e := c.DirectCall(req)
		require.NoError(t, e)
		require.Equal(t, _minimal_service_reply, *(req.Reply.(*string)))
	}

}
