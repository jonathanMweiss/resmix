package config

import (
	"fmt"
	"github.com/jonathanMweiss/resmix/internal/crypto/tibe"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestConfigGeneratePublishersCorrectly(t *testing.T) {
	a := require.New(t)
	hostnames := make([]string, 10)
	for i := range hostnames {
		hostnames[i] = fmt.Sprintf("localhost:808%d", i)
	}
	cnfgs := CreateConfigs(hostnames, 5)

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
