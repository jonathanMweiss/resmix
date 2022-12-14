package ecc

type VerifyingEncoderDecoder interface {
	VerifyingEncoder
	VerifyingDecoder
}

type VerifyingEncoder interface {
	// creates an authenticated decoded of the byte array input
	AuthEncode([]byte) ([][][]byte, error)
	Stop()
}

type VerifyingDecoder interface {
	// verifies that the reconstructed shards were not tempered with
	NewShards() [][][]byte
	AuthReconstruct([][][]byte, int) ([]byte, error)
	Stop()
}
