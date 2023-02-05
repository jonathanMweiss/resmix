package resmix

import (
	"github.com/jonathanMweiss/resmix/internal/crypto/tibe"
)

func createConfigs(addresses []string, polyDegree int) []*ServerConfig {
	shrs, bytePkeys := DKGSetup(addresses, polyDegree)

	polynomials := vsspolynomialsSetup(len(addresses), polyDegree)

	serverConfigs := make([]*ServerConfig, len(addresses))
	for i := range addresses {
		bts, err := shrs[i].Marshal()
		if err != nil {
			panic(err)
		}

		serverConfigs[i] = &ServerConfig{
			Hostname: addresses[i],

			Id: uint32(i),

			SecretDKGShare: bts,
			DKGPublicKeys:  bytePkeys,
			VSSPolynomial:  polynomials[i],
		}
	}
	return serverConfigs
}

func vsspolynomialsSetup(numPolynomials int, polyDegree int) [][]byte {
	polynomials := make([][]byte, numPolynomials)
	for i := 0; i < numPolynomials; i++ {
		poly := tibe.NewRandomPoly(polyDegree)

		bts, err := poly.Marshal()
		if err != nil {
			panic(err)
		}

		polynomials = append(polynomials, bts)
	}

	return polynomials
}

func DKGSetup(addresses []string, d int) ([]tibe.PolyShare, [][]byte) {
	dkgPoly := tibe.NewRandomPoly(d) // so we need half +1 to reconstruct

	shrs := dkgPoly.CreateShares(len(addresses))

	bytePkeys := make([][]byte, len(shrs))
	for i, shr := range shrs {
		bytePkeys[i] = shr.ComputePublicKey().Marshal()
	}

	return shrs, bytePkeys
}
