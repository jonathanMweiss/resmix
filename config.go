package resmix

import (
	"github.com/jonathanMweiss/resmix/internal/crypto/tibe"
)

func createConfigs(addresses []string, polyDegree int) []*ServerConfig {
	dkgShrs, dkgBytePkeys := DKGSetup(addresses, polyDegree)

	serverConfigs := make([]*ServerConfig, len(addresses))
	for i := range addresses {
		bts, err := dkgShrs[i].Marshal()
		if err != nil {
			panic(err)
		}

		serverConfigs[i] = &ServerConfig{

			Hostname:       addresses[i],
			Id:             uint32(i),
			Threshold:      uint32(polyDegree + 1),
			SecretDKGShare: bts,
			DKGPublicKeys:  dkgBytePkeys,
			IBEConfigs: &IBEConfigs{
				VSSPolynomial:              nil,
				VSSExponentPolynomials:     map[string][]byte{},
				AddressOfNodeToSecretShare: map[string][]byte{},
			},
		}
	}

	// setup VSS (n^2): n nodes, each node has n shares
	polys, mPolynomials := vsspolynomialsSetup(len(addresses), polyDegree)

	for i := range serverConfigs {
		serverConfigs[i].IBEConfigs.VSSPolynomial = mPolynomials[i]

		shrs, expoly := tibe.NewNode(polys[i]).VssShares(len(serverConfigs))

		expolyMrshl := expoly.Marshal()

		for j := range serverConfigs {
			serverConfigs[j].IBEConfigs.VSSExponentPolynomials[serverConfigs[i].Hostname] = expolyMrshl

			bts, err := shrs[j].Marshal()
			if err != nil {
				panic(err)
			}

			serverConfigs[j].IBEConfigs.AddressOfNodeToSecretShare[serverConfigs[i].Hostname] = bts
		}
	}

	return serverConfigs
}

func vsspolynomialsSetup(numPolynomials int, polyDegree int) ([]tibe.Poly, [][]byte) {
	mPolynomials := make([][]byte, numPolynomials)
	polynomials := make([]tibe.Poly, numPolynomials)
	for i := 0; i < numPolynomials; i++ {
		polynomials[i] = tibe.NewRandomPoly(polyDegree)

		bts, err := polynomials[i].Marshal()
		if err != nil {
			panic(err)
		}

		mPolynomials = append(mPolynomials, bts)
	}

	return polynomials, mPolynomials
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

func (s *ServerConfig) CreateTIBEPublisher() (tibe.Publisher, error) {
	pkeys := make([]tibe.PublicKey, len(s.DKGPublicKeys))
	for i, bts := range s.DKGPublicKeys {
		if err := pkeys[i].SetBytes(bts); err != nil {
			return nil, err
		}
	}

	privateShare := tibe.PolyShare{}
	if err := privateShare.SetBytes(s.SecretDKGShare); err != nil {
		return nil, err
	}

	return tibe.NewPublisher(int(s.Threshold), privateShare, pkeys)
}

func (s *ServerConfig) CreateTIBENode() (tibe.VssIbeNode, error) {
	cnfg := s.IBEConfigs
	p := tibe.Poly{}

	if err := p.SetBytes(cnfg.VSSPolynomial); err != nil {
		return nil, err
	}

	node := tibe.NewNode(p)

	for hostname, expolyBytes := range cnfg.VSSExponentPolynomials {
		expoly := &tibe.ExponentPoly{}
		if err := expoly.SetBytes(expolyBytes); err != nil {
			return nil, err
		}

		shr := tibe.PolyShare{}
		if err := shr.SetBytes(cnfg.AddressOfNodeToSecretShare[hostname]); err != nil {
			return nil, err
		}

		node.ReceiveShare(hostname, tibe.VssShare{
			ExponentPoly: expoly,
			PolyShare:    shr,
		})
	}

	return node, nil
}
