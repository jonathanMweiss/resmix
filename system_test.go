package resmix

import (
	"context"
	"github.com/jonathanMweiss/resmix/config"
	"github.com/jonathanMweiss/resmix/rrpc"
	"github.com/stretchr/testify/require"
	"net"
	"strconv"
	"testing"
)

type T struct {
	require.TestingT
	sys           *config.SystemConfig
	mixServers    []*server
	rrpcServers   []rrpc.Server
	mixServersMap map[string]*server

	testMsgs TestMessages
	round    int
}

type TestMessages struct {
	mg *MessageGenerator

	Onions     []Onion
	MessageSet map[string]struct{}

	mixToOnions map[string][]Onion
}

func TestSystem(tmp *testing.T) {
	numServers := 10
	polyDegree := 5
	numLayers := 6

	t := newTester(tmp, numServers, polyDegree, numLayers)
	defer t.Close()

	for i := 0; i < 5; i++ {
		tmp.Run("round "+strconv.Itoa(i), func(tt *testing.T) {
			t.setRound(i)

			t.GenerateMessages(100)

			t.startRound()

			t.sendStartMessage()

			response := t.collectMessages()
			verifyResults(t, t.testMsgs.MessageSet, makeSet(response))

			t.endRound()

		})

	}

}

func (t *T) Close() {
	for _, s := range t.mixServers {
		require.NoError(t, s.Close())
	}

	closeServers(t.rrpcServers)
}

func (t *T) setRound(rnd int) {
	t.round = rnd
}

func (t *T) GenerateMessages(n int) {
	if t.testMsgs.mg == nil {
		t.testMsgs.mg = NewMessageGenerator(t.sys)
	}

	pairs, err := t.testMsgs.mg.LoadOrCreateMessages(n, t.round)
	require.NoError(t, err)

	t.testMsgs.Onions = pairs.AllOnions()

	msgs := pairs.AllMessages()
	msgSet := makeSet(msgs)
	require.Equal(t, len(msgSet), len(msgs), "duplicates in message set")

	t.testMsgs.MessageSet = msgSet
	t.testMsgs.mixToOnions = GroupOnionsByMixName(t.testMsgs.Onions, t.sys.Topology)
}

func newTester(t require.TestingT, numServers int, polyDegree int, numLayers int) *T {
	sys := config.CreateLocalSystemConfigs(numServers, polyDegree, numLayers)

	mixServers := createServers(t, sys)

	rrpcServers := launchServers(t, mixServers)

	mixServersMap := mixServerArrayToMap(mixServers)
	return &T{
		TestingT:      t,
		sys:           sys,
		mixServers:    mixServers,
		rrpcServers:   rrpcServers,
		mixServersMap: mixServersMap,
	}
}

func verifyResults(t *T, msgSet map[string]struct{}, responseSet map[string]struct{}) {
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

func (t *T) collectMessages() [][]byte {
	lastLayer := t.sys.Topology.Layers[len(t.sys.Topology.Layers)-1]
	out := make([][]byte, 0, len(t.testMsgs.MessageSet))
	for _, mix := range lastLayer.LogicalMixes {
		resp, err := t.mixServersMap[mix.Hostname].GetMessages(context.Background(), &GetMessagesRequest{
			Round:        uint32(t.round),
			LogicalMixer: []byte(mix.Name),
		})

		require.NoError(t, err)

		out = append(out, resp.Messages...)
	}

	return out
}

func (t *T) startRound() {
	mixToWorkload := make(map[string]uint64)
	for s, onions := range t.testMsgs.mixToOnions {
		mixToWorkload[s] = uint64(len(onions))
	}

	for _, mixServer := range t.mixServers {
		require.NoError(t, mixServer.Dial())

		_, err := mixServer.NewRound(context.Background(), &NewRoundRequest{
			Round:                    uint32(t.round),
			MixIdsToExpectedWorkload: mixToWorkload,
		})

		require.NoError(t, err)
	}
}

func (t *T) endRound() {
	for _, mixServer := range t.mixServers {
		_, err := mixServer.EndRound(context.Background(), &EndRoundRequest{
			Round: uint32(t.round),
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

func (t *T) sendStartMessage() {
	for _, mixServer := range t.mixServers {
		firstMix := mixServer.Configurations.ServerConfig.GetMixesSortedByLayer()[0]
		onions := t.testMsgs.mixToOnions[firstMix]

		_, err := mixServer.AddMessages(context.Background(), &AddMessagesRequest{
			Round: uint32(t.round),
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

func closeServers(servers []rrpc.Server) {
	for _, server := range servers {
		server.Stop()
	}
}

func createServers(t require.TestingT, sys *config.SystemConfig) []*server {
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

func launchServers(t require.TestingT, mixServers []*server) []rrpc.Server {
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
