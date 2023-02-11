package resmix

import (
	"github.com/jonathanMweiss/resmix/config"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestOnions(t *testing.T) {
	numLayers := 10

	sys := config.CreateLocalSystemConfigs(10, 5, numLayers)
	gen := NewMessageGenerator(sys)

	round := 0

	onions := gen.generateOnion([]byte("hello world"), round)
	for _, onion := range onions {
		for i := 0; i < numLayers; i++ {
			mixName := onion.ExtractMixName(sys.Topology)
			hostname := sys.Topology.Mixes[mixName].Hostname
			cnfg := sys.GetServerConfig(hostname)

			nd, err := cnfg.CreateTIBENode()
			require.NoError(t, err)

			cphr := onion.ExtractCipher()
			dc := nd.Decrypter(computeId(hostname, round))
			onion, err = dc.Decrypt(*cphr)
			require.NoError(t, err)
		}

		require.Equal(t, "hello world", string(onion), "onion unwrapping failed")
	}
}

func TestWriteOnionDistribution(t *testing.T) {
	numLayers := 10

	sys := config.CreateLocalSystemConfigs(10, 5, numLayers)
	gen := NewMessageGenerator(sys)

	round := 0
	onions, _ := gen.LoadOrCreateMessagesForClients(100, round)

	onionsPerHostname := make(map[string]int)
	for _, onion := range onions {
		mixname := onion.ExtractMixName(sys.Topology)
		hostname := sys.Topology.Mixes[mixname].Hostname
		onionsPerHostname[hostname] = onionsPerHostname[hostname] + 1
	}

	for host, numOnions := range onionsPerHostname {
		t.Logf("%s: %d", host, numOnions)
	}
}

func BenchmarkOnionGen(b *testing.B) {
	numLayers := 10

	sys := config.CreateLocalSystemConfigs(10, 2, numLayers)
	gen := NewMessageGenerator(sys)

	msg := make([]byte, messageSize)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.generateOnion(msg, 0)
	}
}
