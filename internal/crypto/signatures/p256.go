package signatures

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"github.com/jonathanMweiss/resmix/internal/codec"
	"github.com/jonathanMweiss/resmix/internal/crypto"
	"math/big"
)

// working with ecdsa.p256

type Signature struct {
	R []byte
	S []byte
}

func marshalSignature(r, s *big.Int) ([]byte, error) {
	sig := &Signature{
		R: r.Bytes(),
		S: s.Bytes(),
	}
	return codec.Marshal(sig)
}

func unmarshalSignature(data []byte) (r, s *big.Int, err error) {
	sig := Signature{}
	err = codec.Unmarshal(data, &sig)
	if err != nil {
		return
	}
	r, s = &big.Int{}, &big.Int{}
	r.SetBytes(sig.R)
	s.SetBytes(sig.S)
	return
}

var hash = sha256.Sum256

func hashAndSign(sk *ecdsa.PrivateKey, data []byte) (sig []byte, err error) {
	digest := hash(data)
	r, s, err := ecdsa.Sign(rand.Reader, sk, digest[:])
	if err != nil {
		return
	}
	sig, err = marshalSignature(r, s)
	return
}

func Sign(sk *ecdsa.PrivateKey, v interface{}) (sig []byte, err error) {
	bs, err := codec.Marshal(v)
	if err != nil {
		return nil, err
	}
	return hashAndSign(sk, bs)
}

func Verify(pk *ecdsa.PublicKey, v interface{}, sig []byte) (bool, error) {
	bs, err := codec.Marshal(v)
	if err != nil {
		return false, err
	}
	return hashAndVerify(pk, bs, sig)
}

func WVerify(pk *ecdsa.PublicKey, v interface{}, sig []byte, w crypto.BWriter) (bool, error) {
	start := w.Len()
	if err := codec.MarshalIntoWriter(v, w); err != nil {
		return false, err
	}
	return hashAndVerify(pk, w.Bytes()[start:], sig)
}

func hashAndVerify(pk *ecdsa.PublicKey, data []byte, sig []byte) (bool, error) {
	digest := hash(data)
	r, s, err := unmarshalSignature(sig)
	if err != nil {
		return false, err
	}
	return ecdsa.Verify(pk, digest[:], r, s), nil
}

var BadSignature = fmt.Errorf("bad signature")
