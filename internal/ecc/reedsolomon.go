package ecc

import (
	"bytes"
	"context"
	"fmt"
	"runtime"

	"github.com/klauspost/reedsolomon"
)

const NumChunks = 36

type RSEncoder struct {
	numShards   int
	numParities int
	// the rsEncoder is thread safe according to https://github.com/klauspost/reedsolomon/issues/95
	rsEncoder       reedsolomon.Encoder
	encodeTaskQueue chan encodeTaskRequest
	decodeTaskQueue chan decodeTaskRequest

	// determine the state of this encode to the workers.
	ctx    context.Context
	cancel context.CancelFunc
}

func (rs *RSEncoder) NewShards() [][][]byte {
	shardChunks := make([][][]byte, NumChunks)
	for i := 0; i < NumChunks; i++ {
		shardChunks[i] = make([][]byte, rs.numShards+rs.numParities)
	}
	return shardChunks
}

var errRSEncoderStopped = fmt.Errorf("the reed solomon encoder decoder stopped working")

// used as padding to support tiny messages.
var zeros = make([]byte, NumChunks)

func (rs *RSEncoder) AuthEncode(data []byte) ([][][]byte, error) {
	data = append(data, zeros...)
	// states whether there are workers or not to encode.
	stop := rs.ctx.Done()

	chunkSize := len(data) / NumChunks
	currentAmount := 0
	chunkShards := make([][][]byte, NumChunks)
	responseChan := make(chan encodeTaskResponse, NumChunks) // making certain no worker will get stuck on response.

	for i := 0; i < NumChunks; i++ {
		chunk := data[currentAmount : currentAmount+chunkSize]
		if i+1 >= NumChunks {
			chunk = data[currentAmount:]
		}
		currentAmount += chunkSize

		request := encodeTaskRequest{
			index:        i,
			chunk:        chunk,
			responseChan: responseChan,
		}
		select {
		case rs.encodeTaskQueue <- request:
		case <-stop:
			return nil, errRSEncoderStopped
		}
	}

	for i := 0; i < NumChunks; i++ {
		select {
		case resp := <-responseChan:
			if resp.err != nil {
				return nil, resp.err
			}
			chunkShards[resp.index] = resp.shards
		case <-stop:
			return nil, errRSEncoderStopped
		}
	}
	return chunkShards, nil
}

func (rs *RSEncoder) authEncode(data []byte) ([][]byte, error) {
	dt := make([]byte, len(data))
	copy(dt, data)
	shards, err := rs.rsEncoder.Split(dt)
	if err != nil {
		return nil, err
	}
	err = rs.rsEncoder.Encode(shards)
	if err != nil {
		return nil, err
	}
	return shards, nil
}

type segment struct {
	start int
	end   int
}

func (rs *RSEncoder) AuthReconstruct(shards [][][]byte, dataSize int) ([]byte, error) {
	dataSize = dataSize + NumChunks
	stop := rs.ctx.Done()
	//we get a block of blobs, we need to asses what is the size of each chunk right? yupsy doodle.

	segmentSize := dataSize / NumChunks
	currentAmount := 0
	data := make([]byte, dataSize)
	segments := make([]segment, NumChunks)
	responseChan := make(chan decodeTaskResponse, NumChunks) // making certain no worker will get stuck on response.
	for i := 0; i < NumChunks; i++ {
		chunkSize := segmentSize
		segments[i] = segment{currentAmount, currentAmount + chunkSize}
		if i+1 >= NumChunks {
			chunkSize = dataSize - currentAmount
			segments[i] = segment{currentAmount, dataSize}

		}

		currentAmount += chunkSize

		rqst := decodeTaskRequest{
			size:         chunkSize,
			chunkIndex:   i,
			shards:       shards[i],
			responseChan: responseChan,
		}
		select {
		case rs.decodeTaskQueue <- rqst:
		case <-stop:
			return nil, errRSEncoderStopped
		}
	}

	for i := 0; i < NumChunks; i++ {
		select {
		case rsp := <-responseChan:
			if rsp.err != nil {
				return nil, rsp.err
			}
			s := segments[rsp.chunkIndex]
			copy(data[s.start:s.end], rsp.chunk)
		case <-stop:
			return nil, errRSEncoderStopped
		}
	}

	return data[:len(data)-len(zeros)], nil
}

//todo consider adding a way with a reusable ioBuffer.
func (rs *RSEncoder) authReconstruct(shards [][]byte, dataSize int) ([]byte, error) {
	bf := bytes.NewBuffer(make([]byte, 0, dataSize))
	err := rs.rsEncoder.Reconstruct(shards)
	if err != nil {
		return nil, err
	}

	// adding the size of the hashSum to the actual size.
	err = rs.rsEncoder.Join(bf, shards, dataSize)
	if err != nil {
		return nil, err
	}
	return bf.Bytes(), nil
}

func (rs *RSEncoder) Stop() {
	rs.cancel()
}

func NewRSEncoderDecoder(numShards, NumParities int) (VerifyingEncoderDecoder, error) {
	encoder, err := reedsolomon.New(numShards, NumParities)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())

	rs := &RSEncoder{
		numShards:       numShards,
		numParities:     NumParities,
		rsEncoder:       encoder,
		encodeTaskQueue: make(chan encodeTaskRequest, NumChunks*runtime.NumCPU()),
		decodeTaskQueue: make(chan decodeTaskRequest, NumChunks*runtime.NumCPU()),
		ctx:             ctx,
		cancel:          cancel,
	}
	for i := 0; i < runtime.NumCPU(); i++ {
		go rs.worker()
	}
	return rs, nil
}

type encodeTaskRequest struct {
	index        int
	chunk        []byte
	responseChan chan encodeTaskResponse
}

type encodeTaskResponse struct {
	index  int
	shards [][]byte
	err    error
}

type decodeTaskRequest struct {
	size         int
	chunkIndex   int
	shards       [][]byte
	responseChan chan decodeTaskResponse
}

type decodeTaskResponse struct {
	chunkIndex int
	chunk      []byte
	err        error
}

func (rs RSEncoder) worker() {
	for {
		select {
		case t := <-rs.encodeTaskQueue:
			shards, err := rs.authEncode(t.chunk)
			t.responseChan <- encodeTaskResponse{
				index:  t.index,
				shards: shards,
				err:    err,
			}

		case t := <-rs.decodeTaskQueue:
			data, err := rs.authReconstruct(t.shards, t.size)
			t.responseChan <- decodeTaskResponse{
				chunkIndex: t.chunkIndex,
				chunk:      data,
				err:        err,
			}

		case <-rs.ctx.Done():
			return
		}
	}
}
