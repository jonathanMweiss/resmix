package tibe

import (
	"fmt"
	"github.com/cloudflare/circl/ecc/bls12381"
	"runtime"
	"sync"
)

// Encrypter is the interface for encrypting messages (the master public key of an IBE scheme).
type Encrypter interface {
	Encrypt(ID, msg []byte) (Cipher, error)
}

// Decrypter is the interface for decrypting messages (the private key of an IBE scheme, tied to a specific ID).
type Decrypter interface {
	Decrypt(c Cipher) ([]byte, error)
}

// Master represents the master key and secret key of the IBE scheme.
type Master interface {
	Decrypter(id []byte) Decrypter
	// GetMasterPublicKey returns the master public key of the IBE scheme. The MPK should be used to encrypt messages
	// to any upcoming round-key.
	GetMasterPublicKey() (MasterPublicKey, error)
	VssShares(n int) ([]PolyShare, *ExponentPoly)
}

// VssIbeNode is a node in the VSS IBE scheme. Can receive shares from other Ibe nodes, and reconstruct
// Decrypter for specific shares.
type VssIbeNode interface {
	Master
	// Vote is a vote for a specific ID for a specific VSS share of some other node.
	Vote(otherNodeName string, id []byte) (Vote, error)
	ReceiveShare(otherNodeName string, shr VssShare)
	ReconstructDecrypter(otherNodeName string, votes []Vote) (Decrypter, error)
}

// Cipher is the ciphertext, can be decrypted by a specific Decrypter.
type Cipher struct {
	Gr        *bls12381.G1
	ID        []byte
	Encrypted []byte
}

// MasterPublicKey is an Encrypter.
type MasterPublicKey struct {
	G1 *bls12381.G1
}

// a decryptor tied to a specific ID.
type idBasedPrivateKey struct {
	*bls12381.G2
}

// Encrypt generate a Cipher.
func (p MasterPublicKey) Encrypt(ID, msg []byte) (Cipher, error) {
	r, err := randomScalar()
	if err != nil {
		return Cipher{}, err
	}

	h := g2Hash(ID)
	tmp := &bls12381.G1{}
	tmp.ScalarMult(r, p.G1)

	k, err := hashGt(bls12381.Pair(tmp, h))
	if err != nil {
		return Cipher{}, err
	}

	c, err := aesEncrypt(k, msg)
	if err != nil {
		return Cipher{}, err
	}

	tmp.ScalarMult(r, bls12381.G1Generator())

	return Cipher{
		Gr:        tmp,
		ID:        ID,
		Encrypted: c,
	}, nil
}

// Decrypt decrypts a Cipher.
func (sk idBasedPrivateKey) Decrypt(c Cipher) ([]byte, error) {
	k, err := hashGt(bls12381.Pair(c.Gr, sk.G2))
	if err != nil {
		return nil, err
	}

	return aesDecrypt(k, c.Encrypted)
}

// Node is one that can participate in the TIBE scheme. it should be able to:
// 1. Gen IBE private keys (with its master key).
// 2. Receive shares from other nodes.
// 3. Reconstruct a specific IBE key from a set of votes.
// 4. compute Decrypter tied to specific Ids.
type Node struct {
	*shareableIbeScheme
	Shares    *sync.Map // map[string]VssShare
	workQueue chan validateTask
}

// VssShare holds the share of the underlying polynomial, and the public polynomial (exPoly)
type VssShare struct {
	*ExponentPoly
	PolyShare
}

type validateTask struct {
	*Vote
	*ExponentPoly
	response chan error
}

// NewNode is the constructor for a VssIbeNode.
func NewNode(poly Poly) VssIbeNode {
	ibe := newShareableIbeScheme(poly)
	queue := make(chan validateTask, runtime.NumCPU())

	for i := 0; i < runtime.NumCPU(); i++ {
		go func() {
			for tsk := range queue {
				vote := tsk.Vote
				publicShare := tsk.ExponentPoly.GetPublicShare(uint64(vote.Index))

				if !isValidSignature(vote.ID, vote.Sig, publicShare) {
					tsk.response <- fmt.Errorf("invalid signature: %v", vote)
				}

				close(tsk.response) // sending nil if not full
			}
		}()
	}

	return Node{
		shareableIbeScheme: &ibe,
		Shares:             &sync.Map{},
		workQueue:          queue,
	}
}

// Vote is the proper to share a part of an Decrypter.
func (t Node) Vote(otherNodeName string, id []byte) (Vote, error) {
	shr, ok := t.Shares.Load(otherNodeName)
	if !ok {
		return Vote{}, fmt.Errorf("no share for node %v", otherNodeName)
	}

	sk := shr.(VssShare).PolyShare.Value

	sig := g2Hash(id)
	sig.ScalarMult(sk, sig)

	return Vote{
		ID:    id,
		Index: shr.(VssShare).PolyShare.Index,
		Sig:   sig,
	}, nil
}

// ReceiveShare assumes trusted content.
func (t Node) ReceiveShare(otherNodeName string, shr VssShare) {
	t.Shares.Store(otherNodeName, shr)
}

// ReconstructDecrypter receives a specific node name and a set of votes to reconstruct the Decrypter.
func (t Node) ReconstructDecrypter(nodeName string, votes []Vote) (Decrypter, error) {
	if len(votes) <= 0 {
		return idBasedPrivateKey{}, fmt.Errorf("no votes")
	}

	shr, ok := t.Shares.Load(nodeName)
	if !ok {
		return idBasedPrivateKey{}, fmt.Errorf("no share for node/index %v", votes[0].Index)
	}
	// should verify each vote. (this is why one would need the exPoly.

	expoly := shr.(VssShare).ExponentPoly
	resps := make([]chan error, len(votes))

	for i, vote := range votes {
		vt := vote
		rsp := make(chan error, 1)
		t.workQueue <- validateTask{
			Vote:         &vt,
			ExponentPoly: expoly,
			response:     rsp,
		}

		resps[i] = rsp
	}

	for i := 0; i < len(votes); i++ {
		if err := <-resps[i]; err != nil {
			return idBasedPrivateKey{}, err
		}
	}

	skey, err := reconstructSecretIbeKey(votes, expoly.Threshold())
	if err != nil {
		return idBasedPrivateKey{}, err
	}

	return idBasedPrivateKey{skey}, nil
}

// privates:

type shareableIbeScheme struct {
	P  *Poly
	PK MasterPublicKey
	S  *bls12381.Scalar
}

func (I shareableIbeScheme) VssShares(n int) ([]PolyShare, *ExponentPoly) {
	return I.P.CreateShares(n), I.P.GetExponentPoly()
}

func newShareableIbeScheme(poly Poly) shareableIbeScheme {
	cpyPoly := poly.Copy()
	masterSecret := cpyPoly.Secret()
	masterpub := &bls12381.G1{}
	masterpub.ScalarMult(masterSecret, bls12381.G1Generator())

	return shareableIbeScheme{
		P:  cpyPoly,
		PK: MasterPublicKey{masterpub},
		S:  masterSecret,
	}
}

func (I shareableIbeScheme) Decrypter(id []byte) Decrypter {
	sk := g2Hash(id)
	sk.ScalarMult(I.S, sk)

	return idBasedPrivateKey{sk}
}

func (I shareableIbeScheme) GetMasterPublicKey() (MasterPublicKey, error) {
	mpkCopy := &bls12381.G1{}
	if err := mpkCopy.SetBytes(I.PK.G1.Bytes()); err != nil {
		return MasterPublicKey{}, err
	}

	return MasterPublicKey{G1: mpkCopy}, nil
}
