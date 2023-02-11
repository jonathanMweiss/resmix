package resmix

import (
	"context"
	"github.com/jonathanMweiss/resmix/config"
	"github.com/jonathanMweiss/resmix/rrpc"
	"github.com/stretchr/testify/require"
	"net"
	"testing"
)

func TestSystem(t *testing.T) {
	numServers := 10

	sys := config.CreateLocalSystemConfigs(numServers, 5, 6)

	mixServers := createServers(t, sys)
	defer closeMixServers(mixServers)

	rrpcServers := launchServers(t, mixServers)
	defer closeServers(rrpcServers)

	mixToOnions := genMsgsForTest(t, sys)

	mixToWorkload := make(map[string]uint64)
	for s, onions := range mixToOnions {
		mixToWorkload[s] = uint64(len(onions))
	}

	for _, mixServer := range mixServers {
		require.NoError(t, mixServer.Dial())

		_, err := mixServer.NewRound(context.Background(), &NewRoundRequest{
			Round:                    0,
			MixIdsToExpectedWorkload: mixToWorkload,
		})

		require.NoError(t, err)
	}

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

func onionsToRepeatedByteArrays(onions []Onion) [][]byte {
	res := make([][]byte, len(onions))

	for i, onion := range onions {
		res[i] = onion
	}

	return res
}

func genMsgsForTest(t *testing.T, sys *config.SystemConfig) map[string][]Onion {
	mg := NewMessageGenerator(sys)
	onions, err := mg.LoadOrCreateMessages(1000, 0)
	require.NoError(t, err)

	return GroupOnionsByMixName(onions, sys.Topology)
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
