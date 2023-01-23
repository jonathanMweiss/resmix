package rrpc

import (
	"bytes"
	"fmt"
	"math"
	"sort"

	"github.com/jonathanMweiss/resmix/internal/crypto"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/jonathanMweiss/resmix/internal/ecc"
)

// SemiNetwork internally keeps track of all relay servers and various
// states associated with the networkConfig.
type NetData interface {
	// creates a new error correction code, to be used.
	// notice that creating ecc.VerifyingEncoderDecoder might spin up goroutines.
	NewErrorCorrectionCode() (ecc.VerifyingEncoderDecoder, error)

	// Servers return the list of addresses of the servers.
	Servers() []string

	// maximal errors allowed in the networkConfig.
	MinimalRelayedParcels() int

	// amount of redundancies needed.
	MaxErrors() int

	// Used to state whether we recognise this public key or not.
	ContainKey(key []byte) bool

	// gets a key for the specific ip
	// on err returns a grpc.status err as error
	GetPublicKey(ip string) (crypto.PublicKey, error)

	// gets a hostname from a specific key
	GetHostname(publicKey []byte) string

	//the minimal value needed to generate a proof
	MinimalAttestationNumber() int
}

// Network contains any data used in the rrpc, along with Connections to the relays.
type Network interface {
	NetData

	RelayDial() error
	CloseConnections() error

	PublishProof(*Proof)
	RobustRequest(context context.Context, requests []*RelayRequest) ([]*CallStreamResponse, error)
	CancelRequest(uuid string)
}

type ServerData struct {
	Address   string
	Publickey []byte
}

func (d ServerData) Equal(other ServerData) bool {
	return d.Address == other.Address && bytes.Equal(d.Publickey, other.Publickey)
}

// NetworkConfig represent the configurations that are set in an rRPC system
type NetworkConfig struct {
	Tau           int
	ServerConfigs []ServerData
}

type semiNet struct {
	serverAddresses []string
	// used as a set for keys (string([]byte))
	pkToHost map[string]string
	// maps between hostname and key
	hostToPK map[string]crypto.PublicKey

	Tau int
}
type network struct {
	NetData
	skey  crypto.PrivateKey
	conns map[string]*RelayConn
}

func (n *network) RobustRequest(context context.Context, requests []*RelayRequest) ([]*CallStreamResponse, error) {
	if len(requests) != len(n.conns) {
		return nil, fmt.Errorf("Bad request, number of requests differs from number of relays")
	}
	responseChan := make(chan relayResponse, len(requests))
	srvrs := n.Servers()
	for i := range requests {
		v, ok := n.conns[srvrs[requests[i].Parcel.RelayIndex]]
		if !ok {
			panic("relay index not found")
		}
		v.prepareForRequest(requests)
		v.sendRequest(relayConnRequest{
			RelayRequest: requests[i],
			response:     responseChan,
		})
	}

	responses := make([]*CallStreamResponse, 0, n.MinimalRelayedParcels())
	totalErrors := 0
	var r relayResponse
	var err error
	for range responses {

		select {
		case <-context.Done():
			return nil, context.Err()

		case r = <-responseChan:
		}

		if r.RelayStreamError != nil {
			totalErrors += 1
			if totalErrors >= n.MaxErrors() {
				err = status.ErrorProto(r.RelayStreamError)
				return nil, err
			}
		}

		responses = append(responses, r.Response)
		if len(responses) >= n.MinimalRelayedParcels() {
			return responses, nil
		}
	}
	panic("should never arrive here")
}

func (n *network) CancelRequest(uuid string) {
	for _, conn := range n.conns {
		conn.cancelRequest(uuid)
	}
}

func (n *network) PublishProof(p *Proof) {
	for _, conn := range n.conns {
		conn.SendProof(p)
	}
}

func (n *network) GetRelayConn(hostname string) *RelayConn {
	return n.conns[hostname]
}

func NewNetwork(netdata NetData, skey crypto.PrivateKey) Network {
	return &network{
		NetData: netdata,
		skey:    skey,
		conns:   make(map[string]*RelayConn, len(netdata.Servers())),
	}
}

func (n *network) CloseConnections() error {
	var err error
	for _, conn := range n.conns {
		if err = conn.Close(); err != nil {
			fmt.Println("error closing relay connection:", err)
		}
	}

	return err
}

func (n *network) RelayDial() error {
	for _, s := range n.NetData.Servers() {
		conn, err := NewRelayConn(s)
		if err != nil {
			return n.CloseConnections()
		}

		n.conns[s] = conn
	}

	return nil
}

func (n *semiNet) Servers() []string {
	return n.serverAddresses
}

func (n *semiNet) NewErrorCorrectionCode() (ecc.VerifyingEncoderDecoder, error) {
	maxErasures := n.MaxErrors()
	numDataShards := n.NumServers() - maxErasures
	return ecc.NewRSEncoderDecoder(numDataShards, maxErasures)
}

func (n *semiNet) MinimalAttestationNumber() int {
	// tau > 2(t+x+l)
	// tau/2 > t+x+l
	// hence: tau/2 >= t+x+l+1 which is isMinimalAttestationNumber
	return int(math.Ceil(float64(n.Tau)/2)) + 1 // we add +1 because #attests > t+x+l+1
}

var unknownHostErr = status.Newf(codes.NotFound, "unknown hots").Err()

func (n *semiNet) GetPublicKey(hostname string) (crypto.PublicKey, error) {
	if pk, ok := n.hostToPK[hostname]; ok {
		return pk, nil
	}
	return nil, unknownHostErr
}

func (n *semiNet) ContainKey(key []byte) bool {
	if _, ok := n.pkToHost[string(key)]; ok {
		return true
	}
	return false
}

func (n *semiNet) MaxErrors() int {
	return n.Tau - 1
}

func (n *semiNet) MinimalRelayedParcels() int {
	return n.NumServers() - n.MaxErrors()

}

func (n *semiNet) GetHostname(pub []byte) string {
	return n.pkToHost[string(pub)]
}

func (n *semiNet) NumServers() int {
	return len(n.serverAddresses)
}

// build new networkConfig type
func NewNetData(config *NetworkConfig) NetData {
	nt := &semiNet{
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

func (n *semiNet) addServer(scnf ServerData) {
	n.serverAddresses = append(n.serverAddresses, scnf.Address)
	n.hostToPK[scnf.Address] = scnf.Publickey
	n.pkToHost[string(scnf.Publickey)] = scnf.Address
}

func isvalid(scnf ServerData) bool {
	return scnf.Address != "" && scnf.Publickey != nil
}
