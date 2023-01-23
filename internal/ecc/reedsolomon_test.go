package ecc

import (
	"bytes"
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testCorrectDecodeTiny(t *testing.T, rs VerifyingEncoderDecoder) {
	data := []byte("toosmall")

	shardedMsg, err := rs.AuthEncode(data)
	if err != nil {
		t.Errorf("function failed before decrypt")
	}
	// corrupting message
	for _, sh := range shardedMsg {
		sh[0] = nil
	}

	decodedData, err := rs.AuthReconstruct(shardedMsg, len(data))
	if err != nil {
		t.Errorf("could not decode, Received %v err", err)
		return

	}
	if !bytes.Equal(data, decodedData) {
		t.Errorf("didn't reconstruct correctly, \n %v \n %v", string(data), string(decodedData))
	}
}
func testCorrectDecodeSmall(t *testing.T, rs VerifyingEncoderDecoder) {
	data := []byte("This is a check, there should not be any problem.")

	shardedMsg, err := rs.AuthEncode(data)
	if err != nil {
		t.Errorf("function failed before decrypt")
	}
	// corrupting message
	for _, sh := range shardedMsg {
		sh[0] = nil
	}

	decodedData, err := rs.AuthReconstruct(shardedMsg, len(data))
	if err != nil {
		t.Errorf("could not decode, Received %v err", err)
		return

	}
	if !bytes.Equal(data, decodedData) {
		t.Errorf("didn't reconstruct correctly, \n %v \n %v", string(data), string(decodedData))
	}
}

func testCorrectDecodeWithCorruptions(t *testing.T, rs VerifyingEncoderDecoder) {
	data := make([]byte, 1578)
	if _, err := rand.Read(data); err != nil {
		t.Fatal(err)
	}

	shardedMsg, err := rs.AuthEncode(data)
	if err != nil {
		t.Errorf("function failed before decrypt")
	}

	// corrupting message
	for _, sh := range shardedMsg {
		sh[0] = nil
		sh[1] = nil
		sh[3] = nil
	}

	decodedData, err := rs.AuthReconstruct(shardedMsg, len(data))
	if err != nil {
		t.Errorf("could not decode, Received %v err", err)
		return

	}
	if !bytes.Equal(data, decodedData) {
		t.Errorf("didn't reconstruct correctly, \n %v \n %v", string(data), string(decodedData))
	}
}

func BenchmarkNewRSEncoderDecoder200mb(b *testing.B) {
	rs, _ := NewRSEncoderDecoder(11, 9)
	buff200mb := make([]byte, 1024*1024*200)
	_, _ = rand.Read(buff200mb)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		shards, _ := rs.AuthEncode(buff200mb)
		_, err := rs.AuthReconstruct(shards, 1024*1024*200)
		require.NoError(b, err)
	}
}

func BenchmarkNewRSEncoderDecoderJustEncoding200m(b *testing.B) {
	rs, _ := NewRSEncoderDecoder(11, 9)
	buff200mb := make([]byte, 1024*1024*200)
	_, _ = rand.Read(buff200mb)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err := rs.AuthEncode(buff200mb)
		require.NoError(b, err)
	}
}
func BenchmarkNewRSEncoderDecoderJustDecoding200mb(b *testing.B) {
	rs, _ := NewRSEncoderDecoder(11, 9)
	buff200mb := make([]byte, 1024*1024*200)
	_, _ = rand.Read(buff200mb)

	shards, _ := rs.AuthEncode(buff200mb)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err := rs.AuthReconstruct(shards, 1024*1024*200)
		require.NoError(b, err)
	}
}

func BenchmarkNewRSEncoderDecoder400mb(b *testing.B) {
	rs, _ := NewRSEncoderDecoder(11, 9)
	buff400mb := make([]byte, 1024*1024*400)
	_, _ = rand.Read(buff400mb)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		shards, _ := rs.AuthEncode(buff400mb)
		_, err := rs.AuthReconstruct(shards, 1024*1024*400)
		require.NoError(b, err)
	}
}

func BenchmarkNewRSEncoderDecoderJustEncoding400m(b *testing.B) {
	rs, _ := NewRSEncoderDecoder(11, 9)
	buff200mb := make([]byte, 1024*1024*400)
	_, _ = rand.Read(buff200mb)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err := rs.AuthEncode(buff200mb)
		require.NoError(b, err)
	}
}
func BenchmarkNewRSEncoderDecoderJustDecoding400mb(b *testing.B) {
	rs, _ := NewRSEncoderDecoder(11, 9)
	buff200mb := make([]byte, 1024*1024*400)
	_, _ = rand.Read(buff200mb)

	shards, _ := rs.AuthEncode(buff200mb)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err := rs.AuthReconstruct(shards, 1024*1024*400)
		require.NoError(b, err)
	}
}

func TestRS(t *testing.T) {
	rsTests := []struct {
		description string
		function    func(*testing.T, VerifyingEncoderDecoder)
	}{
		{
			description: "testDecodeSmallItems",
			function:    testCorrectDecodeSmall,
		},
		{
			description: "testCorrectDecodeWithCorruptions",
			function:    testCorrectDecodeWithCorruptions,
		},
		{
			description: "testCorrectDecodeTiny",
			function:    testCorrectDecodeTiny,
		},
		//{
		//description: "corruptDecode",
		//function:    testCorruptDecode,
		//},
	}
	for _, test := range rsTests {
		t.Run(test.description, func(t *testing.T) {
			t.Parallel()
			rs, err := NewRSEncoderDecoder(5, 3)
			assert.NoError(t, err)
			test.function(t, rs)
			rs.Stop()
		})
	}
}
