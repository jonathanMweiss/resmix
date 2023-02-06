package config

import (
	"fmt"
	"github.com/jonathanMweiss/resmix/internal/crypto/tibe"
	"github.com/stretchr/testify/require"
	"testing"
)

func createConfigs(nServers int) []*ServerConfig {
	hostnames := createHostnames(nServers)

	cnfgs := CreateConfigs(hostnames, nServers/2)

	return cnfgs
}

func createHostnames(nServers int) []string {
	hostnames := make([]string, nServers)

	for i := range hostnames {
		hostnames[i] = fmt.Sprintf("localhost:808%d", i)
	}
	return hostnames
}

func TestConfigGeneratePublishersCorrectly(t *testing.T) {
	a := require.New(t)
	cnfgs := createConfigs(10)

	publishers := make([]tibe.Publisher, len(cnfgs))
	for i, cnfg := range cnfgs {
		publisher, err := cnfg.CreateTIBEPublisher()
		a.NoError(err)

		publishers[i] = publisher
	}

	ID := []byte("HELLO")

	secret, err := publishers[0].Hide(ID, []byte("secret message"))
	a.NoError(err)

	votes := make([]tibe.Vote, len(publishers))
	for i, publisher := range publishers {
		votes[i] = publisher.Vote(secret.ID)
	}

	msg, err := publishers[1].Reveal(votes, secret)
	a.NoError(err)
	a.Equal(msg, []byte("secret message"))
}

func TestVSSCreation(t *testing.T) {
	a := require.New(t)
	cnfgs := createConfigs(10)

	masterKeys := make([]tibe.VssIbeNode, len(cnfgs))
	for i, cnfg := range cnfgs {
		masterKey, err := cnfg.CreateTIBENode()
		a.NoError(err)

		masterKeys[i] = masterKey
	}
	mixName := cnfgs[0].Hostname
	id := fmt.Sprintf("%s|round1", mixName)

	// Verify encrypt decrypt
	ctx, err := masterKeys[1].EncryptFor(mixName, []byte(id), []byte("secret message"))
	a.NoError(err)

	ptx, err := masterKeys[0].Decrypter([]byte(id)).Decrypt(ctx)
	a.NoError(err)

	require.Equal(t, ptx, []byte("secret message"))

	votes := make([]tibe.Vote, len(masterKeys)-1)
	// verify key reconstruct from shares:
	for i := 1; i < len(masterKeys); i++ {
		vt, err := masterKeys[i].Vote(mixName, []byte(id))
		a.NoError(err)
		votes[i-1] = vt
	}

	decrypter, err := masterKeys[1].ReconstructDecrypter(mixName, votes)
	a.NoError(err)

	ptx2, err := decrypter.Decrypt(ctx)
	a.NoError(err)

	a.Equal(ptx, ptx2)
}

func TestTopology(t *testing.T) {
	hostnames := createHostnames(5)
	nlayers := 4

	top := CreateCascadeTopology(hostnames, nlayers)
	require.Equal(t, 4, len(top.Layers))

	for _, layer := range top.Layers {
		require.Equal(t, len(hostnames), len(layer.LogicalMixes))
	}

	for i, mix := range top.Layers[0].LogicalMixes {
		require.Equal(t, hostnames[i], mix.Hostname)
		require.Equal(t, i, int(mix.ServerIndex))
		require.Equal(t, 0, int(mix.Layer))

		require.Equal(t, fmt.Sprintf("m(0,%d)", i), mix.Name)
	}

	// validating one predecessor chain.
	currentMix := top.Layers[len(top.Layers)-1].LogicalMixes[0]
	l, idx := 3, 3
	for l != -1 {
		require.Equal(t, fmt.Sprintf("m(%d,%d)", l, idx), currentMix.Name)
		currentMix = currentMix.Predecessors[0]
		l -= 1
		idx -= 1
	}
	require.Equal(t, GenesisName, currentMix.Name)

	currentMix = top.Layers[0].LogicalMixes[0]
	l, idx = 0, 0
	for l != nlayers {
		require.Equal(t, fmt.Sprintf("m(%d,%d)", l, idx), currentMix.Name)
		currentMix = currentMix.Successors[0]
		l += 1
		idx += 1
	}
}
