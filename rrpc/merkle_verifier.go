package rrpc

import (
	"bytes"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jonathanMweiss/resmix/internal/crypto"
	"github.com/jonathanMweiss/resmix/internal/crypto/merklearray"
)

type verifiable interface {
	crypto.Hashable
	popCert() *MerkleCertificate
	pushCert(certificate *MerkleCertificate)
}

//
// signature verification
//
type signatureState struct {
	state *int64
}

func newSigState() *signatureState {
	var state int64 = 0
	return &signatureState{&state}
}

// changes the state indefinitely.
func (t *signatureState) shouldVerifySig() bool {
	// compare to the original value which is 0, if it okay will swap with
	// -1 indicating now that it's no longer available for signing.
	return atomic.CompareAndSwapInt64(t.state, 0, -1)
}

type merklecerttask interface{}

// we might want to share this merkle signer between clients.
type MerkleCertVerifier struct {
	// all verification tasks are passed over this channel
	tasks chan merklecerttask
	// used once the object is to be recycled.
	done    chan bool
	once    sync.Once
	taskMap sync.Map
}

func NewVerifier(numWorkers int) *MerkleCertVerifier {
	verifier := &MerkleCertVerifier{
		taskMap: sync.Map{},
		tasks:   make(chan merklecerttask, numWorkers*8),
		done:    make(chan bool),
	}
	verifier.setWorkers(numWorkers)
	verifier.CreateCleaner()
	return verifier
}

type merkleVerifyTask struct {
	publicKey  crypto.PublicKey
	merkleCert *MerkleCertificate

	// This lock is locked once and this terminate the status of this task. can only attempt lock - so no one waits on the lock once locked.
	signatureState *signatureState
	err            error
	// a channel to state the response is ready
	ready chan interface{}
}

type merkleProofAuthTask struct {
	objectToVerify crypto.Hashable
	merkleCert     *MerkleCertificate
	responseChan   chan error
}

var errEmptyCertificate = errors.New("empty certificate")

func (m *MerkleCertVerifier) Verify(pk []byte, v verifiable) error {
	cert := v.popCert()
	defer v.pushCert(cert)

	if cert == nil {
		return errEmptyCertificate
	}
	var root crypto.Digest
	copy(root[:], cert.Root)

	authTask := &merkleProofAuthTask{
		objectToVerify: v,
		merkleCert:     cert,
		// this is per merkle auth, can't be shared by different verify requests.
		responseChan: make(chan error, 1),
	}
	m.tasks <- authTask
	if err := <-authTask.responseChan; err != nil {
		return err
	}

	t, _ := m.taskMap.LoadOrStore(root.String(), &merkleVerifyTask{
		publicKey:      pk,
		merkleCert:     cert,
		signatureState: newSigState(),

		//sharing the verification result channel
		ready: make(chan interface{}),
		// updated once the verifyTask is ready.
		err: nil,
	})
	task := t.(*merkleVerifyTask)
	shouldVerify := task.signatureState.shouldVerifySig()
	if shouldVerify {
		// if the signature state is not in progress of signing - should send a task to sign it.
		m.tasks <- task
	}

	// a response for the signature should always arrive on this channel.
	<-task.ready
	if task.err != nil {
		// sometimes there might arrive a bad signature, in that case we want to delete.
		// (so we can reinspect the signature. now all that do not posses the same signature as there was in the task:
		// can retry
		m.taskMap.Delete(root.String())
		// because only one of the certs received the chance to sign, all that do not posses the same signature will try to verify again.
		// each time one of goroutines will sign, thus the recursion will end.
		if !shouldVerify && bytes.Equal(task.merkleCert.Signature, cert.Signature) {
			v.pushCert(cert)
			return m.Verify(pk, v)
		}
		return task.err
	}
	return nil
}

func merkleVerificationWorker(verifier *MerkleCertVerifier) {
	buffer := bytes.NewBuffer(nil)
	for {
		select {
		case t := <-verifier.tasks:
			switch v := t.(type) {
			// TODO; instead of having a switchcase, change the v into an interface which has a func that receive a buffer..
			case *merkleVerifyTask:
				verifyMerkleSig(v, buffer)
			case *merkleProofAuthTask:
				authenticateMerklePath(v, buffer)
			}

		case <-verifier.done:
			return
		}
		buffer.Reset()
	}
}

func authenticateMerklePath(t *merkleProofAuthTask, w crypto.BWriter) {
	root, proof, _, leafIndex := unpackMerkleCert(t.merkleCert)
	defer close(t.responseChan) // no use of this channel after message is sent over it.
	if err := merklearray.Verify(root, map[uint64]crypto.Hashable{leafIndex: t.objectToVerify}, proof, w); err != nil {
		t.responseChan <- err
		return
	}
	t.responseChan <- nil
}

func (m *MerkleCertVerifier) clean() {
	// deletes the old map. and releases the memory.
	m.taskMap.Range(func(key, value interface{}) bool {
		m.taskMap.Delete(key)
		return true
	})
}

func verifyMerkleSig(t *merkleVerifyTask, w crypto.BWriter) {
	// broadcasting that the result is ready to anyone waiting on the channel.
	defer close(t.ready)
	root, _, sig, _ := unpackMerkleCert(t.merkleCert)

	validSig, err := t.publicKey.OWVerify(root, sig, w)
	if err != nil {
		t.err = err
		return
	}
	if !validSig {
		t.err = crypto.ErrBadSignature
		return
	}
}

func (m *MerkleCertVerifier) Stop() {
	m.once.Do(func() { close(m.done) })
	m.clean()
}

func (m *MerkleCertVerifier) CreateCleaner() {
	go func() {
		for {
			select {
			case <-m.done:
				return
			case <-time.After(time.Second * 5):
				m.clean()
			}
		}
	}()
}

func (m *MerkleCertVerifier) setWorkers(numWorkeres int) {
	for i := 0; i < numWorkeres; i++ {
		go merkleVerificationWorker(m)
	}
}

func unpackMerkleCert(cert *MerkleCertificate) (
	root crypto.Digest, proof []crypto.Digest, signature []byte, leafIndex uint64) {
	r, path, signature, leafIndex := cert.Root, cert.Path, cert.Signature, cert.Index
	copy(root[:], r)

	proof = make([]crypto.Digest, len(path))

	for i := range path {
		copy(proof[i][:], path[i])
	}

	return
}
