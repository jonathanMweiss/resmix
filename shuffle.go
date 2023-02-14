// This code is a small tweak to the code of XRD repo in github:kwonalbert!

package resmix

import (
	"encoding/binary"
	"io"
)

// Shuffler generates shuffles to onions
type Shuffler struct {
	perm []int
	rand io.Reader
}

// NewShuffler generate a shuffler that randomly shuffles an array
// using rand as the source of randomness
func NewShuffler(rand io.Reader) *Shuffler {
	if rand == nil {
		panic("rand cannot be nil")
	}

	return &Shuffler{
		perm: nil,
		rand: rand,
	}
}

// ShuffleOnions array
func (shf *Shuffler) ShuffleOnions(in []Onion) {
	orig := make([]Onion, len(in))
	copy(orig, in)

	var p []int
	if shf.perm == nil || len(shf.perm) != len(in) {
		shf.permutation(len(in))
	}
	p = shf.perm

	for i := 0; i < len(in); i++ {
		in[i] = orig[p[i]]
	}
}

// generate a random permutation of n elements
func (shf *Shuffler) permutation(n int) {
	arr := make([]int, n)
	for i := 0; i < n; i++ {
		arr[i] = i
	}

	for i := n - 1; i >= 0; i-- {
		j := shf.randInt(i + 1)
		arr[i], arr[j] = arr[j], arr[i]
	}

	shf.perm = arr
}

// generate a random integer < mod
func (shf *Shuffler) randInt(mod int) int {
	tmp := make([]byte, 4)
	_, _ = shf.rand.Read(tmp) // ignoring random's error.
	val := int(binary.BigEndian.Uint32(tmp))
	return val % mod
}
