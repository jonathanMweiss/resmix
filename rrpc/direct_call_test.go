package rrpc

import (
	"context"
	"fmt"
	"github.com/jonathanMweiss/resmix/internal/crypto"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net"
	"strconv"
	"testing"
)

const _serverport = "600"
const _minimal_service_reply = "minimal service reply"

var simplereply Services = Services{
	"testService": {
		Server: (new)(bool),
		MethodDescriptors: map[string]*MethodDesc{
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
	networks   []Coordinator
}

func (c *clientTestSetup) start(t require.TestingT) {
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

func newClientTestSetup(t require.TestingT, srvc Services) clientTestSetup {

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
	netdata := NewNetworkData(netconf)

	networks := make([]Coordinator, 0, len(netdata.Servers()))
	srvrs := make([]RrpcServer, 0, len(netdata.Servers()))

	for i, s := range netdata.Servers() {
		l, err := net.Listen("tcp", s)
		require.NoError(t, err)

		network := NewCoordinator(netdata, sks[i])
		networks = append(networks, network)

		srvr, err := newServerService(srvc, network)
		require.NoError(t, err)

		srvrs = append(srvrs, srvr)

		go func() {
			if srvr.Serve(l) != nil {
				panic("server failed to serve")
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

func TestDirectCallWithStructService(t *testing.T) {
	type simpleService struct {
		message string
	}

	service := simpleService{
		message: "testService",
	}

	services := Services{
		"testService": {
			ServerType: (new)(simpleService),
			Server:     service,
			MethodDescriptors: map[string]*MethodDesc{
				"testMethod": {
					Name: "testMethod",
					Handler: func(server interface{}, ctx context.Context, dec func(interface{}) error) (interface{}, error) {

						return server.(simpleService).message, nil
					},
				},
			},
		},
	}

	setup := newClientTestSetup(t, services)

	setup.start(t)
	defer setup.releaseResources()

	// Ensuring the coordinator dials to all relays.
	c := newClient(setup.serverAddr, setup.networks[0])
	defer c.Close()

	req := &Request{
		Args:    nil,
		Reply:   new(string),
		Method:  "testService/testMethod",
		Uuid:    "1234",
		Context: context.Background(),
	}

	e := c.DirectCall(req)
	require.NoError(t, e)
	require.Equal(t, service.message, *(req.Reply.(*string)))
}

func TestDirectCall(t *testing.T) {
	setup := newClientTestSetup(t, simplereply)

	setup.start(t)
	defer setup.releaseResources()

	// Ensuring the coordinator dials to all relays.
	c := newClient(setup.serverAddr, setup.networks[0])
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

func TestDirectCallErrors(t *testing.T) {
	var testerr = status.Error(codes.DataLoss, "test error")
	var service Services = Services{
		"testService": {
			Server: (new)(bool),
			MethodDescriptors: map[string]*MethodDesc{
				"testMethod": {
					Name: "testMethod",
					Handler: func(server interface{}, ctx context.Context, dec func(interface{}) error) (interface{}, error) {
						return _minimal_service_reply, testerr
					},
				},
			},
		},
	}

	setup := newClientTestSetup(t, service)

	setup.start(t)
	defer setup.releaseResources()

	// Ensuring the coordinator dials to all relays.
	c := newClient(setup.serverAddr, setup.networks[0])
	defer c.Close()

	req := &Request{
		Args:    nil,
		Reply:   new(string),
		Method:  "testService/testMethod",
		Uuid:    "1234",
		Context: context.Background(),
	}

	e := c.DirectCall(req)
	require.Error(t, e)
	require.Equal(t, codes.DataLoss, status.Code(e))
}

func BenchmarkDirectCall(b *testing.B) {
	setup := newClientTestSetup(b, simplereply)

	setup.start(b)
	defer setup.releaseResources()

	// Ensuring the coordinator dials to all relays.
	c := newClient(setup.serverAddr, setup.networks[0])

	req := &Request{
		Args:    nil,
		Reply:   new(string),
		Method:  "testService/testMethod",
		Uuid:    "0",
		Context: context.Background(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req.Uuid = strconv.Itoa(i)
		require.NoError(b, c.DirectCall(req))
	}
	b.StopTimer()
}
