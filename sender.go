package resmix

import (
	"context"
	"fmt"
	"github.com/jonathanMweiss/resmix/config"
	"github.com/jonathanMweiss/resmix/rrpc"
	"github.com/sirupsen/logrus"
)

type Sender interface {
	Update(recoveryScheme)
	Close()
}

// TODO: should have access to the mixesoutputs. that means that it should know about all mixes that had already sent,
// 	and attempt resending them if necessary.

type messageSender struct {
	round int

	topology    *config.Topology
	connections map[hostname]rrpc.ClientConn
	input       <-chan mixoutput
	address     []byte

	log *logrus.Entry
}

func (m messageSender) Close() {
	// TODO:
}
func onionsToRepeatedByteArrays(onions []Onion) [][]byte {
	res := make([][]byte, len(onions))

	for i, onion := range onions {
		res[i] = onion
	}

	return res
}

func (m *messageSender) worker() {
	for tosend := range m.input {
		// group by hostname
		groupedOnions := GroupOnionsByHostname(tosend.onions, m.topology)

		for host, onionsByHost := range groupedOnions {
			rq := m.createRRPCRequest(onionsByHost, tosend, host)

			if err := m.connections[hostname(host)].DirectCall(rq); err != nil {
				m.log.Errorln("direct call failed: ", err.Error())
			}
		}
	}
}

func (m *messageSender) createRRPCRequest(onionsByHost []Onion, tosend mixoutput, host string) *rrpc.Request {
	mp := GroupOnionsByMixName(onionsByHost, m.topology)

	msgs := &AddMessagesRequest{
		Round:    uint32(m.round),
		Messages: make([]*Messages, len(mp)),
	}

	i := 0

	for mixToSendTo, onions := range mp {
		msgs.Messages[i] = &Messages{
			Messages:        onionsToRepeatedByteArrays(onions),
			PhysicalSender:  m.address,
			LogicalSender:   tosend.logicalSender,
			LogicalReceiver: []byte(mixToSendTo),
		}

		i += 1
	}

	rq := &rrpc.Request{
		Args:    msgs,
		Reply:   &AddMessagesResponse{},
		Method:  "/resmix.Mix/AddMessages",
		Uuid:    fmt.Sprintf("%v-to-%v:from-%v", m.address, host, string(tosend.logicalSender)),
		Context: context.Background(), // TODO: get valid timeout for this context.
	}
	return rq
}

func NewSender(round int, h string, topology *config.Topology, connections map[hostname]rrpc.ClientConn, input <-chan mixoutput) *messageSender {
	s := &messageSender{
		round:       round,
		topology:    topology,
		connections: connections,
		input:       input,
		address:     []byte(h),
		log:         logrus.WithFields(logrus.Fields{"component": "mix.Sender", "address": h, "round": round}),
	}

	go s.worker()
	return s
}
