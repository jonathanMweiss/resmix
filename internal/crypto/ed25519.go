package crypto

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"fmt"

	"github.com/jonathanMweiss/resmix/internal/codec"
)

type PrivateKey ed25519.PrivateKey
type PublicKey ed25519.PublicKey

var ErrBadSignature = fmt.Errorf("bad signature")

func GenerateKeys() (PrivateKey, PublicKey, error) {
	p, s, e := ed25519.GenerateKey(rand.Reader)
	if e != nil {
		return nil, nil, e
	}
	return PrivateKey(s), PublicKey(p), nil
}

func (s PrivateKey) Public() PublicKey {
	return PublicKey(ed25519.PrivateKey(s).Public().(ed25519.PublicKey))
}

func (s PrivateKey) Sign(msg []byte) []byte {
	return ed25519.Sign((ed25519.PrivateKey)(s), msg)
}

// OWSign signs any object by marshaling it first into bytecode using a buffer, and then signs it.
func (s PrivateKey) OWSign(v interface{}, w BWriter) ([]byte, error) {
	start := w.Len()
	if err := codec.MarshalIntoWriter(v, w); err != nil {
		return nil, err
	}
	return s.Sign(w.Bytes()[start:]), nil
}

func (p PublicKey) Verify(sig, msg []byte) bool {
	return ed25519.Verify(ed25519.PublicKey(p), msg, sig)
}

// OWVerify is similar to OVerify, but optimised to write into a buffer with the wanted capacity.
func (p PublicKey) OWVerify(v interface{}, sig []byte, w BWriter) (bool, error) {
	start := w.Len()
	if err := codec.MarshalIntoWriter(v, w); err != nil {
		return false, err
	}
	return p.Verify(sig, w.Bytes()[start:]), nil
}

func (p PublicKey) Equal(other PublicKey) bool {
	return bytes.Equal(p, other)
}

// Encoding decoding:

func (p PublicKey) Marshal() []byte {
	return p
}

func (s PrivateKey) Marshal() []byte {
	return s
}

func DecodePKey(encodedPubKey []byte) (PublicKey, error) {
	if len(encodedPubKey) != ed25519.PublicKeySize {
		return nil, errors.New(
			fmt.Sprintf("bad key size, expected:%v received: %v", ed25519.PublicKeySize, len(encodedPubKey)),
		)
	}
	return encodedPubKey, nil
}

func DecodeSKey(encodedPrivateKey []byte) (PrivateKey, error) {
	if len(encodedPrivateKey) != ed25519.PrivateKeySize {
		return nil, errors.New(
			fmt.Sprintf(
				"bad key size, expected:%v received: %v", ed25519.PrivateKeySize, len(encodedPrivateKey),
			),
		)
	}
	return encodedPrivateKey, nil
}
