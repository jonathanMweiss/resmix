package tibe

import (
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"
)

func TestVSSIBEEncryptDecrypt(t *testing.T) {
	p := NewRandomPoly(5)
	vb := newShareableIbeScheme(p)
	sk := vb.Decrypter([]byte("key1"))
	c, err := vb.PK.Encrypt([]byte("key1"), []byte("check"))
	require.NoError(t, err)

	o, err := sk.Decrypt(c)
	require.NoError(t, err)
	require.Equal(t, o, []byte("check"))
}

func TestVSSIBEReconstructKeyForSpecificIDUsingNodes(t *testing.T) {
	nodes := []VssIbeNode{}
	for i := 0; i < 10; i++ {
		p := NewRandomPoly(5)
		nodes = append(nodes, NewNode(p))
	}

	// send shares:
	for i, node := range nodes {
		shrs, err := node.VssShares(len(nodes))
		require.NoError(t, err)

		for j, ibeNode := range nodes {
			ibeNode.ReceiveShare(strconv.Itoa(i), shrs[j])
		}
	}
	keyID := []byte("key1")
	votes := make([]Vote, len(nodes))
	for i, node := range nodes {
		v, err := node.Vote("0", keyID)
		require.NoError(t, err)
		votes[i] = v
	}

	// reconstruct the key:
	dec, err := nodes[3].ReconstructDecrypter("0", votes)
	require.NoError(t, err)
	// comparing keys:
	require.Equal(t, dec.(idBasedPrivateKey).Bytes(), nodes[0].Decrypter(keyID).(idBasedPrivateKey).Bytes())

	// ensuring both can decrypt the same things:
	mk, err := nodes[0].GetMasterPublicKey()
	require.NoError(t, err)
	c, err := mk.Encrypt(keyID, []byte("check"))
	require.NoError(t, err)
	require.NotEqual(t, c.Encrypted, []byte("check")) // ensuring the ciphertext is not equal to the plaintext.

	msg, err := dec.Decrypt(c)
	require.NoError(t, err)
	msg2, err := nodes[0].Decrypter(keyID).Decrypt(c)
	require.NoError(t, err)
	require.Equal(t, msg, msg2)

}
func TestVSSIBEReconstructKeyForSpecificID(t *testing.T) {
	p := NewRandomPoly(5)
	vb := newShareableIbeScheme(p)

	shrs, err := vb.VssShares(10)
	require.NoError(t, err)

	// create Votes using the shares to get a specific key:
	keyname := []byte("key1")

	votes := make([]Vote, len(shrs))
	for _, shr := range shrs {
		tmp := g2Hash(keyname)
		tmp.ScalarMult(shr.Value, tmp)
		votes[shr.Index-1] = Vote{shr.Index, keyname, tmp}
	}

	c, err := vb.PK.Encrypt(keyname, []byte("check"))
	require.NoError(t, err)

	nd := NewNode(p)
	nd.ReceiveShare("0", VssShare{
		ExponentPoly: shrs[0].ExponentPoly,
		PolyShare:    PolyShare{},
	})
	sk, err := nd.ReconstructDecrypter("0", votes)
	require.NoError(t, err)

	msg, err := sk.Decrypt(c)
	require.NoError(t, err)
	require.Equal(t, msg, []byte("check"))
}

func BenchmarkReconstructIbeDecryptor(b *testing.B) {
	p := NewRandomPoly(50)
	vb := newShareableIbeScheme(p)

	shrs, err := vb.VssShares(100)
	require.NoError(b, err)

	// create Votes using the shares to get a specific key:
	keyname := []byte("key1")

	votes := make([]Vote, len(shrs))
	for _, shr := range shrs {
		tmp := g2Hash(keyname)
		tmp.ScalarMult(shr.Value, tmp)
		votes[shr.Index-1] = Vote{shr.Index, keyname, tmp}
	}

	for _, vote := range votes {
		shrs[0].ExponentPoly.GetPublicShare(uint64(vote.Index))
	}

	nd := NewNode(p)
	nd.ReceiveShare("0", VssShare{
		ExponentPoly: shrs[0].ExponentPoly,
		PolyShare:    PolyShare{},
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := nd.ReconstructDecrypter("0", votes)
		require.NoError(b, err)
	}
}
