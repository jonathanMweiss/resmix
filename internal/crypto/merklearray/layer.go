//copied from Algorand

//Copyright (C) 2019-2020 Algorand, Inc.
// This file is part of go-algorand
//
// go-algorand is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// go-algorand is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with go-algorand.  If not, see <https://www.gnu.org/licenses/>.

package merklearray

import (
	"github.com/jonathanMweiss/resmix/internal/crypto"
)

// A layer of the Merkle tree consists of a dense array of hashes at that
// level of the tree.  Hashes beyond the end of the array (e.g., if the
// number of leaves is not an exact power of 2) are implicitly zero.
type layer []crypto.Digest

// A pair represents an internal node in the Merkle tree.
type pair struct {
	l crypto.Digest
	r crypto.Digest
}

func (p *pair) ToBeHashed(w crypto.BWriter) (crypto.HashID, []byte) {
	start := w.Len()
	w.Write(p.r[:])
	w.Write(p.l[:])
	return crypto.MerkleArrayNode, w.Bytes()[start:] //buf[:]
}

// up takes a layer representing some level in the tree,
// and returns the next-higher level in the tree,
// represented as a layer.
func (l layer) up(w crypto.BWriter) layer {
	res := make(layer, (uint64(len(l))+1)/2)
	for i := 0; i < len(l); i += 2 {
		var p pair
		p.l = l[i]
		if i+1 < len(l) {
			p.r = l[i+1]
		}
		res[i/2] = crypto.HashObj(&p, w)
	}
	return res
}
