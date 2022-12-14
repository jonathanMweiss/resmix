package tibe

import (
	"crypto/rand"
	"fmt"
	"github.com/cloudflare/circl/ecc/bls12381"
	"sort"
	"sync"
)

// Poly is a polynomial of degree n.
type Poly struct {
	Coefs []*bls12381.Scalar
}

// PolyShare is a specific share of a polynomial P(i).
type PolyShare struct {
	Index int
	Value *bls12381.Scalar
}

// NewRandomPoly returns a new polynomial with a random secret and random coefficients.
func NewRandomPoly(degree int) Poly {
	coefs := make([]*bls12381.Scalar, degree+1)
	for i := range coefs {
		ai := &bls12381.Scalar{}
		if err := ai.Random(rand.Reader); err != nil {
			panic(err)
		}

		coefs[i] = ai
	}

	return Poly{Coefs: coefs}
}

// Degree states the degree of the polynomial.
func (p *Poly) Degree() int {
	return len(p.Coefs)
}

// Threshold states the minimum number of shares required to reconstruct the secret.
func (p *Poly) Threshold() int {
	return len(p.Coefs) + 1
}

// Coeffs returns a copy of the coefficients of the polynomial.
func (p *Poly) Coeffs() []*bls12381.Scalar {
	return p.Copy().Coefs
}

// Secret returns the secret value of the polynomial, that is P(0).
func (p *Poly) Secret() *bls12381.Scalar {
	s := &bls12381.Scalar{}
	s.Set(p.Coefs[0])

	return s
}

// Eval Evaluates the polynomial P for a given i: P(i).
func (p *Poly) Eval(i int) *bls12381.Scalar {
	xi := &bls12381.Scalar{}
	xi.SetUint64(uint64(i))

	v := &bls12381.Scalar{}
	v.SetUint64(0)

	for j := len(p.Coefs) - 1; j >= 0; j-- {
		v.Mul(v, xi)
		v.Add(v, p.Coefs[j])
	}

	return v
}

// CreateShares creates a set of shares for the polynomial. these shares are secret values
// and should be treated with care.
func (p *Poly) CreateShares(n int) []PolyShare {
	shrs := make([]PolyShare, n)
	for j := 1; j < n+1; j++ {
		shrs[j-1] = PolyShare{
			Index: j,
			Value: p.Eval(j),
		}
	}

	return shrs
}

// GetExponentPoly returns the poly in an exponent form: [a0,a1,a2,...] turns into [g^a0, g^a1, g^a2,...]
func (p *Poly) GetExponentPoly() *ExponentPoly {
	ep := &ExponentPoly{
		Coefs: make([]*bls12381.G1, p.Degree()),
		cache: sync.Map{},
	}

	for i := range ep.Coefs {
		ep.Coefs[i] = &bls12381.G1{}
		ep.Coefs[i].ScalarMult(p.Coefs[i], bls12381.G1Generator())
	}

	return ep
}

// Copy returns a copy of the polynomial.
func (p *Poly) Copy() *Poly {
	cpy := &Poly{make([]*bls12381.Scalar, len(p.Coefs))}
	for i := range p.Coefs {
		cpy.Coefs[i] = &bls12381.Scalar{}
		cpy.Coefs[i].Set(p.Coefs[i])
	}

	return cpy
}

type lagrangeIndexMultiplier struct {
	I      int // index of holder, 0 <= I < N. will compute \ell_0
	Degree int // degree of polynomial
	Value  *bls12381.Scalar
}

func newLagrangeIndexMultiplier(degree, index int, indices []int) (*lagrangeIndexMultiplier, error) {
	xs := map[int]struct{}{}
	for _, i := range indices {
		xs[i] = struct{}{}
	}

	if len(xs) < degree+1 {
		return nil, fmt.Errorf("missing %d points", degree+1-len(xs))
	}

	inds := make([]int, 0, len(xs))
	for ind := range xs {
		inds = append(inds, ind)
	}

	sort.Ints(inds)
	inds = inds[:degree+1]

	return &lagrangeIndexMultiplier{
		I:      index,
		Degree: degree,
		Value:  computeLagrangeIndexMultiplier(inds, index),
	}, nil
}

func reconstructSecret(xy map[int]*bls12381.Scalar) *bls12381.Scalar {
	nom := &bls12381.Scalar{}
	denom := &bls12381.Scalar{}

	acc := &bls12381.Scalar{}
	acc.SetUint64(0)

	tmp := &bls12381.Scalar{}
	tmpidx := &bls12381.Scalar{}

	for xi, yi := range xy {
		nom.Set(yi)
		denom.SetUint64(1)
		tmp.SetUint64(uint64(xi))

		for xj := range xy {
			if xj == xi {
				continue
			}

			tmpidx.SetUint64(uint64(xj)) // using xj-0 as nominator.
			nom.Mul(nom, tmpidx)

			tmpidx.Sub(tmpidx, tmp)
			denom.Mul(denom, tmpidx)
		}

		denom.Inv(denom)
		nom.Mul(nom, denom)
		acc.Add(acc, nom)
	}

	return acc
}

func computeLagrangeIndexMultiplier(indices []int, index int) *bls12381.Scalar {
	nom := &bls12381.Scalar{}
	nom.SetUint64(1)

	denom := &bls12381.Scalar{}
	denom.SetUint64(1)

	indx := &bls12381.Scalar{}
	indx.SetUint64(uint64(index))

	tmp := &bls12381.Scalar{} // created once, reused throughout the for-loop.
	for _, xj := range indices {
		if xj == index {
			continue
		}

		tmp.SetUint64(uint64(xj)) // using xj-0 as nominator.
		nom.Mul(nom, tmp)

		tmp.Sub(tmp, indx)
		denom.Mul(denom, tmp)
	}

	// nom/denom:
	denom.Inv(denom)
	nom.Mul(nom, denom)

	return nom
}
