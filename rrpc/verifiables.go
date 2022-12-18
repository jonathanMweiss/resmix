package rrpc

import (
	"bytes"
	"fmt"

	"github.com/jonathanMweiss/resmix/internal/codec"
	"github.com/jonathanMweiss/resmix/internal/crypto"
	"github.com/jonathanMweiss/resmix/internal/crypto/merklearray"
)

// MerkleCertifiable represents anything that can be signed using a merkle signature scheme.
type MerkleCertifiable interface {
	verifiable
	SetMerkleCert(root crypto.Digest, proof []crypto.Digest, leafIndex int, signature []byte)
}

// follows the interface Merkle.Array.
type array []MerkleCertifiable

func (a array) Length() uint64 {
	return uint64(len(a))
}

func (a array) Get(pos uint64) (crypto.Hashable, error) {
	length := a.Length()
	if length == 0 {
		return nil, fmt.Errorf("empty array")
	}
	if pos >= length {
		return nil, fmt.Errorf("pos %d larger than length %d", pos, length)
	}
	return a[pos], nil
}

func merkleSign(arr []MerkleCertifiable, sk crypto.PrivateKey) error {
	bf := bytes.NewBuffer(make([]byte, 0, len(arr)*crypto.DigestSize))
	tree, err := merklearray.Build((array)(arr), bf)
	if err != nil {
		return err
	}
	r := tree.Root()
	sig, err := sk.OSign(r)
	if err != nil {
		return err
	}
	bf.Reset()
	for i := range arr {
		proof, err := tree.Prove([]uint64{uint64(i)}, bf)
		if err != nil {
			return err
		}
		arr[i].SetMerkleCert(r, proof, i, sig)
	}
	return nil
}

func proofIntoBytes(proof []crypto.Digest) [][]byte {
	byteProof := make([][]byte, len(proof))
	for i, p := range proof {
		p := p
		byteProof[i] = p[:]
	}
	return byteProof
}

func prepareForHashing(m interface{}, w crypto.BWriter) (crypto.HashID, []byte) {
	start := w.Len()
	if err := codec.MarshalIntoWriter(m, w); err != nil {
		// TODO maybe panic
		return "", nil
	}
	return crypto.Message, w.Bytes()[start:]
}

// ====
// slow requests
// ====

//func (p *Parcel) SetMerkleCert(root crypto.Digest, proof []crypto.Digest, leafIndex int, signature []byte) {
//	p.Merkle = &pb.MerkleCertificate{
//		Root:      root[:],
//		Path:      proofIntoBytes(proof),
//		Index:     uint64(leafIndex),
//		Signature: signature,
//	}
//}

//func (p *parcel) ToBeHashed(w crypto.BWriter) (crypto.HashID, []byte) {
//	return prepareForHashing(p, w)
//}

//func (p *parcel) popCert() *pb.MerkleCertificate {
//	cert := p.Merkle
//	p.Merkle = nil
//	return cert
//}
//
//func (p *parcel) pushCert(cert *pb.MerkleCertificate) {
//	p.Merkle = cert
//}
//
//type relayReq pb.RelayRequest
//
//func (r *relayReq) SetMerkleCert(root crypto.Digest, proof []crypto.Digest, leafIndex int, signature []byte) {
//	(*parcel)(r.Parcel).SetMerkleCert(root, proof, leafIndex, signature)
//}
//
//func (r *relayReq) ToBeHashed(w crypto.BWriter) (crypto.HashID, []byte) {
//	start := w.Len()
//	if err := codec.MarshalIntoWriter(r.Parcel, w); err != nil {
//		// TODO maybe panic
//		return "", nil
//	}
//	return crypto.Message, w.Bytes()[start:]
//}

//type serverResp pb.CallStreamResponse

//func (r *serverResp) SetMerkleCert(root crypto.Digest, proof []crypto.Digest, leafIndex int, signature []byte) {
//	r.Merkle = &pb.MerkleCertificate{
//		Root:      root[:],
//		Path:      proofIntoBytes(proof),
//		Index:     uint64(leafIndex),
//		Signature: signature,
//	}
//}
//
//func (s *serverResp) ToBeHashed(w crypto.BWriter) (crypto.HashID, []byte) {
//	return prepareForHashing(s, w)
//}
//
//func (s *serverResp) popCert() *pb.MerkleCertificate {
//	cert := s.Merkle
//	s.Merkle = nil
//	return cert
//}
//
//func (s *serverResp) pushCert(cert *pb.MerkleCertificate) {
//	s.Merkle = cert
//}

// ====
// fast requests
// ====

// ===
// WorkNotes
// ===

type senderNote ExchangeNote

func (m *senderNote) popCert() *MerkleCertificate {
	cert := m.SenderMerkleProof
	m.SenderMerkleProof = nil
	return cert
}

func (m *senderNote) pushCert(certificate *MerkleCertificate) {
	m.SenderMerkleProof = certificate
}

func (m *senderNote) ToBeHashed(w crypto.BWriter) (crypto.HashID, []byte) {
	return prepareForHashing(m, w)
}

func (m *senderNote) SetMerkleCert(root crypto.Digest, proof []crypto.Digest, leafIndex int, signature []byte) {
	m.SenderMerkleProof = &MerkleCertificate{
		Root:      root[:],
		Path:      proofIntoBytes(proof),
		Index:     uint64(leafIndex),
		Signature: signature,
	}
}

type receiverNote ExchangeNote

func (r *receiverNote) popCert() *MerkleCertificate {
	cert := r.ReceiverMerkleProof
	r.ReceiverMerkleProof = nil
	return cert
}

func (r *receiverNote) pushCert(certificate *MerkleCertificate) {
	r.ReceiverMerkleProof = certificate
}

func (r *receiverNote) ToBeHashed(w crypto.BWriter) (crypto.HashID, []byte) {
	return prepareForHashing(r, w)
}

func (r *receiverNote) SetMerkleCert(root crypto.Digest, proof []crypto.Digest, leafIndex int, signature []byte) {
	r.ReceiverMerkleProof = &MerkleCertificate{
		Root:      root[:],
		Path:      proofIntoBytes(proof),
		Index:     uint64(leafIndex),
		Signature: signature,
	}
}
