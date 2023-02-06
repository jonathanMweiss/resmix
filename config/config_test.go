package config

import (
	"fmt"
	"github.com/jonathanMweiss/resmix/internal/crypto/tibe"
	"github.com/stretchr/testify/require"
	"testing"
)

func createConfigs(nServers int) []*ServerConfig {
	hostnames := make([]string, nServers)

	for i := range hostnames {
		hostnames[i] = fmt.Sprintf("localhost:808%d", i)
	}

	cnfgs := CreateConfigs(hostnames, nServers/2)

	return cnfgs
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
