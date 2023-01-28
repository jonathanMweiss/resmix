package rrpc

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRobustCall(t *testing.T) {
	setup := newClientTestSetup(t)

	setup.start(t)
	defer setup.releaseResources()

	// Ensuring the network dials to all relays.
	c := NewClient(setup.sk, setup.serverAddr, setup.networks[0])
	req := &Request{
		Args:    nil,
		Reply:   new(string),
		Method:  "testService/testMethod",
		Uuid:    "1234",
		Context: context.Background(),
	}
	require.NoError(t, c.RobustCall(req))
	require.Equal(t, _minimal_service_reply, *(req.Reply.(*string)))
	time.Sleep(time.Second)
}
