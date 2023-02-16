package resmix

import (
	"context"
	"github.com/jonathanMweiss/resmix/config"
	"github.com/jonathanMweiss/resmix/rrpc"
	"github.com/stretchr/testify/require"
	"net"
	"testing"
)

type T struct {
	*testing.T
	sys           *config.SystemConfig
	mixServers    []*server
	rrpcServers   []rrpc.Server
	mixServersMap map[string]*server

	testMsgs TestMessages
	round    int
}

type TestMessages struct {
	mg *MessageGenerator

	Onions   []Onion
	Messages [][]byte

	mixToOnions map[string][]Onion
}

func (i *T) Close() {
	closeMixServers(i.mixServers)
	closeServers(i.rrpcServers)
}

func (i *T) setRound(rnd int) {
	i.round = rnd
}

func (i *T) GenerateMessages(n int) {
	if i.testMsgs.mg == nil {
		i.testMsgs.mg = NewMessageGenerator(i.sys)
	}

	pairs, err := i.testMsgs.mg.LoadOrCreateMessages(n, i.round)
	require.NoError(i, err)

	i.testMsgs.Onions = pairs.AllOnions()
	i.testMsgs.Messages = pairs.AllMessages()

	i.testMsgs.mixToOnions = GroupOnionsByMixName(i.testMsgs.Onions, i.sys.Topology)
}

func TestSystem(tmp *testing.T) {
	numServers := 10
	polyDegree := 5
	numLayers := 6

	t := newTester(tmp, numServers, polyDegree, numLayers)
	defer t.Close()

	t.setRound(0)

	t.GenerateMessages(100)
	mixToOnions := t.testMsgs.mixToOnions
	msgs := t.testMsgs.Messages

	t.startRound()

	sendStartMessage(t.T, t.mixServers, mixToOnions)

	msgSet := makeSet(msgs)
	require.Equal(t.T, len(msgSet), len(msgs), "duplicates in message set")

	response := collectMessages(t.T, t.sys, t.mixServersMap)

	verifyResults(t.T, msgSet, makeSet(response))

	endRound(t.T, t.mixServers)
}

func newTester(t *testing.T, numServers int, polyDegree int, numLayers int) *T {
	sys := config.CreateLocalSystemConfigs(numServers, polyDegree, numLayers)

	mixServers := createServers(t, sys)

	rrpcServers := launchServers(t, mixServers)

	mixServersMap := mixServerArrayToMap(mixServers)
	return &T{
		T:             t,
		sys:           sys,
		mixServers:    mixServers,
		rrpcServers:   rrpcServers,
		mixServersMap: mixServersMap,
	}
}

func verifyResults(t *testing.T, msgSet map[string]struct{}, responseSet map[string]struct{}) {
	require.Equal(t, len(msgSet), len(responseSet))
	for msg := range msgSet {
		_, ok := responseSet[msg]
		require.True(t, ok)
	}
}

func makeSet(msgs [][]byte) map[string]struct{} {
	msgMap := make(map[string]struct{}, len(msgs))
	for _, msg := range msgs {
		msgMap[string(msg)] = struct{}{}
	}
	return msgMap
}

func collectMessages(t *testing.T, sys *config.SystemConfig, mixServersMap map[string]*server) [][]byte {
	lastLayer := sys.Topology.Layers[len(sys.Topology.Layers)-1]
	out := make([][]byte, 0, 1000)
	for _, mix := range lastLayer.LogicalMixes {
		resp, err := mixServersMap[mix.Hostname].GetMessages(context.Background(), &GetMessagesRequest{
			Round:        0,
			LogicalMixer: []byte(mix.Name),
		})

		require.NoError(t, err)

		out = append(out, resp.Messages...)
	}

	return out
}

func (i *T) startRound() {
	mixToWorkload := make(map[string]uint64)
	for s, onions := range i.testMsgs.mixToOnions {
		mixToWorkload[s] = uint64(len(onions))
	}

	for _, mixServer := range i.mixServers {
		require.NoError(i, mixServer.Dial())

		_, err := mixServer.NewRound(context.Background(), &NewRoundRequest{
			Round:                    0,
			MixIdsToExpectedWorkload: mixToWorkload,
		})

		require.NoError(i, err)
	}
}

func endRound(t *testing.T, mixServers []*server) {
	for _, mixServer := range mixServers {
		_, err := mixServer.EndRound(context.Background(), &EndRoundRequest{
			Round: 0,
		})

		require.NoError(t, err)
	}
}

func mixServerArrayToMap(servers []*server) map[string]*server {
	mp := make(map[string]*server)
	for _, server := range servers {
		mp[server.Configurations.ServerConfig.Hostname] = server
	}

	return mp
}

func sendStartMessage(t *testing.T, mixServers []*server, mixToOnions map[string][]Onion) {
	for _, mixServer := range mixServers {
		firstMix := mixServer.Configurations.ServerConfig.GetMixesSortedByLayer()[0]
		onions := mixToOnions[firstMix]

		_, err := mixServer.AddMessages(context.Background(), &AddMessagesRequest{
			Round: 0,
			Messages: []*Messages{
				{
					Messages:        onionsToRepeatedByteArrays(onions),
					PhysicalSender:  []byte(config.GenesisName),
					LogicalSender:   []byte(config.GenesisName),
					LogicalReceiver: []byte(firstMix),
				},
			},
		})

		require.NoError(t, err)
	}
}

func closeMixServers(servers []*server) {
	// todo
}

func closeServers(servers []rrpc.Server) {
	for _, server := range servers {
		server.Stop()
	}
}

func createServers(t *testing.T, sys *config.SystemConfig) []*server {
	mixServers := make([]*server, len(sys.ServerConfigs))

	for i, serverConfig := range sys.ServerConfigs {
		msrvr, err := NewMixServer(&ResmixConfigs{
			ServerConfig: serverConfig,
			Topology:     sys.GetTopology(),
			RrpcConfigs: rrpc.Configs{
				IsGrpc:            false,
				ServerCoordinator: serverConfig.CreateCoordinator(),
			},
		})

		require.NoError(t, err)

		mixServers[i] = msrvr.(*server)
	}
	return mixServers
}

func launchServers(t *testing.T, mixServers []*server) []rrpc.Server {
	rrpcSrvrs := make([]rrpc.Server, len(mixServers))

	for i, mixServer := range mixServers {
		srvr, err := rrpc.NewServer(
			rrpc.SingleServiceFromGRPCServiceDesc(_Mix_serviceDesc, MixServer(mixServer)),
			mixServer.Configurations.RrpcConfigs,
		)

		require.NoError(t, err)

		rrpcSrvrs[i] = srvr
		l, err := net.Listen("tcp", mixServer.Configurations.ServerConfig.GetHostname())
		require.NoError(t, err)

		go srvr.Serve(l)
	}
	return rrpcSrvrs
}
