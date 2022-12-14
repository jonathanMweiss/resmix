package tibe

import (
	"github.com/cloudflare/circl/ecc/bls12381"
	"strconv"
	"sync"
)

// ExponentPoly is a public broadcast of a polynomial value. [g^{S}, g^{a_1}, g^{a_2}, ..., g^{a_n}]
// it can be used to verify VSS shares, and to calculate the public share (key) of a player in a VSS protocol.
type ExponentPoly struct {
	Coefs []*bls12381.G1
	// GetPublicShare is an expensive computation, to avoid that we cache the results.
	cache sync.Map
}

// Threshold returns the threshold of the underlying polynomial.
func (e *ExponentPoly) Threshold() int {
	return len(e.Coefs) + 1
}

// GetPublicShare is responsible for calculating the value g^P(i) for the underlying Polynomial.
// g^P(i) = g^a_0 + g^a_1 * i + g^a_2 * i^2 + ... + g^a_n * i^n which can be used as the public key for the secret P(i).
func (e *ExponentPoly) GetPublicShare(i uint64) *bls12381.G1 {
	if i == 0 {
		return e.Coefs[0]
	}

	if v, ok := e.cache.Load(i); ok {
		return v.(*bls12381.G1)
	}

	f := func() {
		v := &bls12381.Scalar{}
		v.SetUint64(i)

		vCopy := &bls12381.Scalar{}
		vCopy.SetUint64(i)

		res := &bls12381.G1{}
		res.SetIdentity()

		// eval the poly at i:
		for i := 1; i < len(e.Coefs); i++ {
			coef := &bls12381.G1{}
			coef.ScalarMult(v, e.Coefs[i])
			res.Add(res, coef)
			// advance to the next exponent...
			v.Mul(v, vCopy)
		}
		res.Add(e.Coefs[0], res)
		e.cache.Store(i, res)
	}

	doOnceKey := strconv.FormatUint(i, 10)
	// ensuring this computation will be used only once.
	o, _ := e.cache.LoadOrStore(doOnceKey, &sync.Once{})

	once, ok := o.(*sync.Once)
	if !ok {
		panic("failed to cast to sync.Once")
	}

	once.Do(f)

	v, ok := e.cache.Load(i)
	if !ok {
		panic("ExponentPoly.GetPublicShare: couldn't create public share")
	}

	return v.(*bls12381.G1)
}

// VerifyShare verifies that a share is valid for the given exponent polynomial.. that is:
// g^P(i) = g^a_0 + g^a_1 * i + g^a_2 * i^2
func (e *ExponentPoly) VerifyShare(index uint64, share *bls12381.Scalar) bool {
	actual := &bls12381.G1{}
	actual.ScalarMult(share, bls12381.G1Generator())

	return actual.IsEqual(e.GetPublicShare(index))
}
