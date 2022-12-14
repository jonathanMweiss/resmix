package signatures

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

func EncodePublicKey(publicKey *ecdsa.PublicKey) []byte {
	x509EncodedPub, _ := x509.MarshalPKIXPublicKey(publicKey)
	return pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: x509EncodedPub})
}

func EncodePrivKey(privateKey *ecdsa.PrivateKey) []byte {
	x509Encoded, _ := x509.MarshalECPrivateKey(privateKey)
	return pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: x509Encoded})
}

func DecodePubKey(encodedPubKey []byte) (*ecdsa.PublicKey, error) {
	// todo understand what to do with the '_'
	blockPub, _ := pem.Decode(encodedPubKey)
	if blockPub == nil {
		return nil, fmt.Errorf("could not decode public key")
	}
	x509EncodedPub := blockPub.Bytes
	genericPublicKey, err := x509.ParsePKIXPublicKey(x509EncodedPub)
	if err != nil {
		return nil, err
	}
	publicKey := genericPublicKey.(*ecdsa.PublicKey)
	return publicKey, nil
}
func DecodePrivateKey(encodedPrivateKey []byte) (*ecdsa.PrivateKey, error) {
	blockPub, _ := pem.Decode(encodedPrivateKey)
	if blockPub == nil {
		return nil, fmt.Errorf("could not decode private key")
	}
	x509EncodedPub := blockPub.Bytes
	privateKey, err := x509.ParseECPrivateKey(x509EncodedPub)
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}
