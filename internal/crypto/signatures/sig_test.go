package signatures

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"testing"

	"github.com/jonathanMweiss/resmix/internal/crypto"
	"github.com/stretchr/testify/require"
)

//136-141.09 microseconds
func BenchmarkEcdsa(b *testing.B) {
	a := require.New(b)
	ecdsakey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	a.NoError(err)
	pk := &ecdsakey.PublicKey
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		buffer := make([]byte, 1000)
		rand.Reader.Read(buffer)
		b.StartTimer()
		sig, err := Sign(ecdsakey, buffer)
		a.NoError(err)
		Verify(pk, buffer, sig)
	}
}

//94.5 microseconds
func BenchmarkEd25519(b *testing.B) {
	a := require.New(b)
	pk, sk, err := ed25519.GenerateKey(rand.Reader)
	a.NoError(err)

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		buffer := make([]byte, 1000)
		rand.Reader.Read(buffer)
		b.StartTimer()
		sig := ed25519.Sign(sk, buffer)
		ed25519.Verify(pk, buffer, sig)

	}
}

//97.83 microseconds -- 100micro.
func BenchmarkMyEd25519(b *testing.B) {
	a := require.New(b)
	s, p, e := crypto.GenerateKeys()
	a.NoError(e)

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		buffer := make([]byte, 1000)
		rand.Reader.Read(buffer)

		b.StartTimer()
		sig, e := s.OSign(buffer)
		a.NoError(e)
		_, e = p.OVerify(sig, buffer)
		a.NoError(e)
	}
}
