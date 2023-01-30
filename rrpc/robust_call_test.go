package rrpc

import (
	"context"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRobustCall(t *testing.T) {
	setup := newClientTestSetup(t, simplereply)

	setup.start(t)
	defer setup.releaseResources()

	// Ensuring the network dials to all relays.
	c := NewClient(setup.sk, setup.serverAddr, setup.networks[0])
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
			server: (new)(bool),
			methodDescriptors: map[string]*MethodDesc{
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

	// Ensuring the network dials to all relays.
	c := NewClient(setup.sk, setup.serverAddr, setup.networks[0])
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

func TestRobustCallFail(t *testing.T) {
	var errorForTest = fmt.Errorf("error for test")
	var lateReply = Services{
		"testService": {
			server: (new)(bool),
			methodDescriptors: map[string]*MethodDesc{
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

	// Ensuring the network dials to all relays.
	c := NewClient(setup.sk, setup.serverAddr, setup.networks[0])
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
