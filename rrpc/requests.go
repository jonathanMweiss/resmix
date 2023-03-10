package rrpc

import (
	"bytes"
	"context"

	"github.com/jonathanMweiss/resmix/internal/codec"
)

type Request struct {
	Args  interface{}
	Reply interface{}
	// metadata to use in the request
	Method string
	Uuid   string

	marshaledArgs  []byte
	marshaledReply []byte
	isUnpacked     bool
	context.Context
}

func (r *Request) pack() error {
	bf := bytes.NewBuffer(make([]byte, 0, 1024))
	return r.packWithBuffer(bf)
}

func (r *Request) unpack() error {
	if r.isUnpacked {
		return nil
	}

	if r.IsCancelled() {
		return context.Canceled
	}

	if err := codec.Unmarshal(r.marshaledReply, r.Reply); err != nil {
		return err
	}

	r.isUnpacked = true

	return nil
}

func (r *Request) packWithBuffer(bf *bytes.Buffer) error {
	if r.marshaledArgs != nil {
		return nil
	}

	if r.IsCancelled() {
		return context.Canceled
	}

	if err := codec.MarshalIntoWriter(r.Args, bf); err != nil {
		return err
	}

	r.marshaledArgs = bf.Bytes()

	return nil
}

func (r *Request) IsCancelled() bool {
	select {
	case <-r.Done():
		return true
	default:
		return false
	}
}
