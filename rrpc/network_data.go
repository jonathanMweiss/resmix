package rrpc

import (
	"bytes"
	"github.com/jonathanMweiss/resmix/internal/crypto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"math"
	"sort"
)

type NetworkData interface {
	Servers() []string

	MinimalRelayedParcels() int

	MaxErrors() int

	ContainKey(key crypto.PublicKey) bool

	GetPublicKey(ip string) (crypto.PublicKey, error)

	GetHostname(publicKey []byte) string

	MinimalAttestationNumber() int

	GetRelayIndex(ip string) int
}

type ServerData struct {
	Address   string
	Publickey []byte
}

func NewNetworkData(config *NetworkConfig) *network {
	nt := &network{
		pkToHost: make(map[string]string),
		hostToPK: make(map[string]crypto.PublicKey),
		Tau:      config.Tau,
	}

	for _, scnf := range config.ServerConfigs {
		if !isvalid(scnf) {
			continue
		}
		nt.addServer(scnf)
	}

	// ensuring everyone holds the same index per server
	sort.Strings(nt.serverAddresses)

	return nt
}

func (d ServerData) Equal(other ServerData) bool {
	return d.Address == other.Address && bytes.Equal(d.Publickey, other.Publickey)
}

// NetworkConfig represent the configurations that are set in an rRPC system
type NetworkConfig struct {
	Tau           int
	ServerConfigs []ServerData
}

type network struct {
	serverAddresses []string
	// used as a set for keys (string([]byte))
	pkToHost map[string]string
	// maps between hostname and key
	hostToPK map[string]crypto.PublicKey

	Tau int
}

func (n *network) GetRelayIndex(ip string) int {
	return sort.SearchStrings(n.serverAddresses, ip) // like searching a tree...
}
func (n *network) Servers() []string {
	return n.serverAddresses
}

func (n *network) MinimalAttestationNumber() int {
	// tau > 2(t+x+l)
	// tau/2 > t+x+l
	// hence: tau/2 >= t+x+l+1 which is isMinimalAttestationNumber
	return int(math.Ceil(float64(n.Tau)/2)) + 1 // we add +1 because #attests > t+x+l+1
}

var unknownHostErr = status.Newf(codes.NotFound, "unknown hots").Err()

func (n *network) GetPublicKey(hostname string) (crypto.PublicKey, error) {
	if pk, ok := n.hostToPK[hostname]; ok {
		return pk, nil
	}
	return nil, unknownHostErr
}

func (n *network) ContainKey(key crypto.PublicKey) bool {
	if _, ok := n.pkToHost[string(key)]; ok {
		return true
	}
	return false
}

func (n *network) MaxErrors() int {
	return n.Tau - 1
}

func (n *network) MinimalRelayedParcels() int {
	return n.NumServers() - n.MaxErrors()

}

func (n *network) GetHostname(pub []byte) string {
	return n.pkToHost[string(pub)]
}

func (n *network) NumServers() int {
	return len(n.serverAddresses)
}

func (n *network) addServer(scnf ServerData) {
	n.serverAddresses = append(n.serverAddresses, scnf.Address)
	n.hostToPK[scnf.Address] = scnf.Publickey
	n.pkToHost[string(scnf.Publickey)] = scnf.Address
}

func isvalid(scnf ServerData) bool {
	return scnf.Address != "" && scnf.Publickey != nil
}
