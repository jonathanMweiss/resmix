package rrpc

import (
	"context"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRobustCall(t *testing.T) {
	setup := newClientTestSetup(t, simplereply)

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
	require.NoError(t, c.RobustCall(req))
	require.Equal(t, _minimal_service_reply, *(req.Reply.(*string)))
}

func TestRobustCallTimeouts(t *testing.T) {
	var lateReply = Services{
		"testService": {
			Server: (new)(bool),
			MethodDescriptors: map[string]*MethodDesc{
				"testMethod": {
					Name: "testMethod",
					Handler: func(server interface{}, ctx context.Context, dec func(interface{}) error) (interface{}, error) {
						time.Sleep(ttl + time.Millisecond*100)
						return _minimal_service_reply, nil
					},
				},
			},
		},
	}

	setup := newClientTestSetup(t, lateReply)

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

	err := c.RobustCall(req)
	require.Error(t, err)

	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, st.Code(), codes.Canceled)
}

func TestRobustCallErrors(t *testing.T) {
	var errorForTest = fmt.Errorf("error for test")
	var lateReply = Services{
		"testService": {
			Server: (new)(bool),
			MethodDescriptors: map[string]*MethodDesc{
				"testMethod": {
					Name: "testMethod",
					Handler: func(server interface{}, ctx context.Context, dec func(interface{}) error) (interface{}, error) {
						return _minimal_service_reply, errorForTest
					},
				},
			},
		},
	}

	setup := newClientTestSetup(t, lateReply)

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

	err_ := c.RobustCall(req)
	require.Error(t, err_)

	require.True(t, strings.Contains(err_.Error(), errorForTest.Error()))
}

func TestRobustCallFailWithSpecificCode(t *testing.T) {
	var errorForTest = status.Error(codes.DataLoss, "error for test")
	var lateReply = Services{
		"testService": {
			Server: (new)(bool),
			MethodDescriptors: map[string]*MethodDesc{
				"testMethod": {
					Name: "testMethod",
					Handler: func(server interface{}, ctx context.Context, dec func(interface{}) error) (interface{}, error) {
						return _minimal_service_reply, errorForTest
					},
				},
			},
		},
	}

	setup := newClientTestSetup(t, lateReply)

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

	err_ := c.RobustCall(req)
	require.Error(t, err_)

	require.True(t, strings.Contains(err_.Error(), errorForTest.Error()))
	require.Equal(t, codes.DataLoss, status.Code(err_))
}

func TestRobustCallWithStructService(t *testing.T) {
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

	e := c.RobustCall(req)
	require.NoError(t, e)
	require.Equal(t, service.message, *(req.Reply.(*string)))
}

func BenchmarkRobustCall(b *testing.B) {
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
		require.NoError(b, c.RobustCall(req))
	}
	b.StopTimer()
	time.Sleep(time.Second * 2)
}
