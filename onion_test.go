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
			mix := onion.ExtractMixConfig(sys.Topology)

			cnfg := sys.GetServerConfig(mix.Hostname)

			nd, err := cnfg.CreateTIBENode()
			require.NoError(t, err)

			cphr := onion.ExtractCipher()
			dc := nd.Decrypter(computeId(mix.Hostname, round))
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
	msgs := MessageOnionsPairArray(gen.generateOnions(100, round))

	grps := GroupOnionsByMixName(msgs.AllOnions(), sys.Topology)

	for host, split := range grps {
		t.Logf("%s: %d", host, len(split))
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

func BenchmarkSingleOnionLayerDecryption(b *testing.B) {
	numLayers := 10

	sys := config.CreateLocalSystemConfigs(10, 5, numLayers)
	gen := NewMessageGenerator(sys)

	round := 0

	onions := gen.generateOnion([]byte("hello world"), round)

	onion := onions[0]

	cps := make([]Onion, b.N)
	for i := 0; i < b.N; i++ {
		cps[i] = make(Onion, len(onion))
		copy(cps[i], onion)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// setup shouldn't be counted - the mix does this on the round setup, not for every message
		b.StopTimer()
		onion := cps[i]
		mix := onion.ExtractMixConfig(sys.Topology)

		cnfg := sys.GetServerConfig(mix.Hostname)

		nd, err := cnfg.CreateTIBENode()
		require.NoError(b, err)

		// timing onion decryption:
		b.StartTimer()
		cphr := onion.ExtractCipher()
		dc := nd.Decrypter(computeId(mix.Hostname, round))
		_, err = dc.Decrypt(*cphr)
		require.NoError(b, err)
	}
}
