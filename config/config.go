package config

import (
	"github.com/jonathanMweiss/resmix/internal/crypto"
	"github.com/jonathanMweiss/resmix/internal/crypto/tibe"
	"github.com/jonathanMweiss/resmix/rrpc"
	"golang.org/x/crypto/sha3"
	"sort"
	"strconv"
)

const seed = "there is nothing up my sleeve"

var randomReader sha3.ShakeHash

func init() {
	cpy := make([]byte, len(seed))

	copy(cpy, seed)

	randomReader = sha3.NewShake128()

	randomReader.Write(cpy)
}

// CreateSystemConfigs
// polyDegree is the degree of the polynomial used for DKG and VSS. determines the threshold of VSS (degree + 1).
func CreateSystemConfigs(addresses []string, polyDegree, numLayers int) *SystemConfig {
	top := CreateCascadeTopology(addresses, numLayers)
	srvrs := createConfigs(addresses, polyDegree)
	for _, srvr := range srvrs {
		srvr.setMixNames(top)
	}

	peers := make([]*Peer, len(srvrs))
	for i, srvr := range srvrs {
		peers[i] = &Peer{
			Hostname:      srvr.GetHostname(),
			RrpcPublicKey: srvr.GetRrpcPublicKey(),
		}
	}

	for _, srvr := range srvrs {
		srvr.Peers = peers
	}

	return &SystemConfig{
		ServerConfigs: srvrs,
		Topology:      top,
	}
}

func CreateLocalSystemConfigs(numServers, polyDegree, numLayers int) *SystemConfig {
	addresses := make([]string, numServers)
	for i := range addresses {
		addresses[i] = "localhost:" + strconv.Itoa(5050+i)
	}

	return CreateSystemConfigs(addresses, polyDegree, numLayers)
}

func createConfigs(addresses []string, polyDegree int) []*ServerConfig {
	dkgShrs, dkgBytePkeys := DKGSetup(addresses, polyDegree)

	serverConfigs := make([]*ServerConfig, len(addresses))
	for i := range addresses {
		bts, err := dkgShrs[i].Marshal()
		if err != nil {
			panic(err)
		}

		sk, pk, err := crypto.GenerateKeys()
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
				VSSPolynomial:                   nil,
				VSSExponentPolynomials:          map[string][]byte{},
				AddressOfNodeToSecretShare:      map[string][]byte{},
				AddressOfNodeToMasterPublicKeys: map[string][]byte{},
			},
			Mixes:         nil, // filled later.
			RrpcSecretKey: sk,
			RrpcPublicKey: pk,
		}
	}

	// setup VSS (n^2): n nodes, each node has n shares
	polys, mPolynomials := vsspolynomialsSetup(len(addresses), polyDegree)

	for i := range serverConfigs {
		serverConfigs[i].IBEConfigs.VSSPolynomial = mPolynomials[i]

		shrs, err := tibe.NewNode(polys[i]).VssShares(len(serverConfigs))
		if err != nil {
			panic(err)
		}

		expolyMrshl := shrs[0].ExponentPoly.Marshal()

		for j := range serverConfigs {
			serverConfigs[j].IBEConfigs.VSSExponentPolynomials[serverConfigs[i].Hostname] = expolyMrshl

			bts, err := shrs[j].PolyShare.Marshal()
			if err != nil {
				panic(err)
			}

			serverConfigs[j].IBEConfigs.AddressOfNodeToSecretShare[serverConfigs[i].Hostname] = bts

			serverConfigs[j].IBEConfigs.AddressOfNodeToMasterPublicKeys[serverConfigs[i].Hostname] = shrs[j].MasterPublicKey.Bytes()
		}
	}

	return serverConfigs
}

func vsspolynomialsSetup(numPolynomials int, polyDegree int) ([]tibe.Poly, [][]byte) {
	mPolynomials := make([][]byte, numPolynomials)
	polynomials := make([]tibe.Poly, numPolynomials)
	for i := 0; i < numPolynomials; i++ {
		polynomials[i] = tibe.NewRandomPoly(polyDegree, randomReader)

		bts, err := polynomials[i].Marshal()
		if err != nil {
			panic(err)
		}

		mPolynomials[i] = bts
	}

	return polynomials, mPolynomials
}

func DKGSetup(addresses []string, d int) ([]tibe.PolyShare, [][]byte) {
	dkgPoly := tibe.NewRandomPoly(d, randomReader) // so we need half +1 to reconstruct

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

		mpk := tibe.MasterPublicKey{}
		if err := mpk.SetBytes(cnfg.AddressOfNodeToMasterPublicKeys[hostname]); err != nil {
			return nil, err
		}

		node.ReceiveShare(hostname, tibe.VssShare{
			ExponentPoly:    expoly,
			PolyShare:       shr,
			MasterPublicKey: mpk,
		})
	}

	return node, nil
}

func (s *ServerConfig) setMixNames(top *Topology) {
	for nm, mix := range top.Mixes {
		if mix.Hostname != s.Hostname {
			continue
		}

		s.Mixes = append(s.Mixes, nm)
	}

	sort.Slice(s.Mixes, func(i, j int) bool {
		return top.Mixes[s.Mixes[i]].Layer < top.Mixes[s.Mixes[j]].Layer
	})
}

func (s *ServerConfig) CreateCoordinator() rrpc.ServerCoordinator {
	ncnfgs := &rrpc.NetworkConfig{
		Tau:           len(s.Peers) / 2,
		ServerConfigs: make([]rrpc.ServerData, len(s.Peers)),
	}

	for i, peer := range s.Peers {
		ncnfgs.ServerConfigs[i] = rrpc.ServerData{
			Address:   peer.Hostname,
			Publickey: peer.RrpcPublicKey,
		}
	}

	return rrpc.NewCoordinator(rrpc.NewNetworkData(ncnfgs), s.RrpcSecretKey)
}

func (s *SystemConfig) GetServerConfig(hostname string) *ServerConfig {
	for _, srv := range s.ServerConfigs {
		if srv.Hostname == hostname {
			return srv
		}
	}

	return nil
}

func (s *ServerConfig) GetMixesSortedByLayer() []string {
	return s.Mixes
}
