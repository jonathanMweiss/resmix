package rrpc

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"testing"

	"github.com/jonathanMweiss/resmix/internal/crypto"
	"github.com/stretchr/testify/require"
)

const _serverport = "600"
const _minimal_service_reply = "minimal service reply"

var simplereply Services = Services{
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
	sk         crypto.PrivateKey
	serverAddr string
	srvrs      []RrpcServer
	networks   []Network
}

func (c *clientTestSetup) start(t *testing.T) {
	for _, n := range c.networks {
		require.NoError(t, n.Dial())
	}
}

func (c *clientTestSetup) releaseResources() {

	fmt.Println("closing servers")
	for i, srvr := range c.srvrs {
		if err := c.networks[i].Close(); err != nil {
			panic(err)
		}

		srvr.Stop()
	}
	fmt.Println("closed servers")
}

func newClientTestSetup(t *testing.T, srvc Services) clientTestSetup {
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

	networks := make([]Network, 0, len(netdata.Servers()))
	srvrs := make([]RrpcServer, 0, len(netdata.Servers()))

	for i, s := range netdata.Servers() {
		l, err := net.Listen("tcp", s)
		require.NoError(t, err)

		network := NewNetwork(netdata, sks[i])
		networks = append(networks, network)

		srvr, err := NewServerService(sks[i], srvc, network)
		require.NoError(t, err)

		srvrs = append(srvrs, srvr)

		go func() {
			if srvr.Serve(l) != nil {
				panic("Server failed to serve")
			}
		}()
	}

	return clientTestSetup{
		networks:   networks,
		sk:         sks[0],
		serverAddr: serverAddr,
		srvrs:      srvrs,
	}
}

func TestDirectCall(t *testing.T) {
	setup := newClientTestSetup(t, simplereply)

	setup.start(t)
	defer setup.releaseResources()

	// Ensuring the network dials to all relays.
	c := NewClient(setup.sk, setup.serverAddr, setup.networks[0])
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
