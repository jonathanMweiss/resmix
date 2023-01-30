package rrpc

import (
	"bytes"
	"fmt"
	"math"
	"runtime"
	"sort"

	"github.com/jonathanMweiss/resmix/internal/crypto"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/jonathanMweiss/resmix/internal/ecc"
)

// SemiNetwork internally keeps track of all relay servers and various
// states associated with the networkConfig.
type NetworkData interface {
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

	GetRelayIndex(ip string) int
}

type RelayGroup interface {
	CancelRequest(uuid string)
	RobustRequest(context context.Context, requests []*RelayRequest) ([]*CallStreamResponse, error)
	PublishProof(*Proof)
}

// Coordinator contains any data used in the rrpc, along with Connections to the relays.
type Coordinator interface {
	NetworkData

	getErrorCorrectionCode() ecc.VerifyingEncoderDecoder
	getVerifier() *MerkleCertVerifier
	getRelayGroup() RelayGroup

	Dial() error
	Close() error
}

type ServerCoordinator interface {
	Coordinator

	AsyncSend(publickey crypto.PublicKey, msg *CallStreamRequest) error

	// Incoming returns anything(!) even timeouts that came over the
	// coordinator, from any of the servers this coordinator is listening on!
	Incoming() <-chan *CallStreamResponse
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

type coordinator struct {
	NetworkData
	*MerkleCertVerifier

	skey        crypto.PrivateKey
	relayConns  map[string]*RelayConn
	serverConns map[string]*ServerConn

	callResponseChan chan *CallStreamResponse

	myAddress string

	ctx    context.Context
	cancel context.CancelFunc
	ecc    ecc.VerifyingEncoderDecoder
}

func (n *coordinator) getErrorCorrectionCode() ecc.VerifyingEncoderDecoder {
	return n.ecc
}

func (n *coordinator) getVerifier() *MerkleCertVerifier {
	return n.MerkleCertVerifier
}

func (n *coordinator) AsyncSend(publickey crypto.PublicKey, msg *CallStreamRequest) error {
	hostname := n.GetHostname(publickey)
	if hostname == "" {
		return status.Error(codes.NotFound, "public key not found")
	}

	conn, ok := n.serverConns[hostname]
	if !ok {
		return status.Error(codes.Internal, "no connection to server")
	}

	conn.send(msg)
	return nil
}

func (n *coordinator) Incoming() <-chan *CallStreamResponse {
	return n.callResponseChan
}

func (n *coordinator) getRelayGroup() RelayGroup {
	return n
}

func (n *coordinator) RobustRequest(ctx context.Context, requests []*RelayRequest) ([]*CallStreamResponse, error) {
	if len(requests) != len(n.relayConns) {
		return nil, fmt.Errorf("bad request, number of requests differs from number of relays")
	}

	responseChan := make(chan relayResponse, len(requests))
	srvrs := n.Servers()

	for i := range requests {
		v, ok := n.relayConns[srvrs[requests[i].Parcel.RelayIndex]]
		if !ok {
			panic("relay index not found")
		}

		v.sendRequest(relayConnRequest{
			RelayRequest: requests[i],
			response:     responseChan,
		})
	}

	responses := make([]*CallStreamResponse, 0, n.MinimalRelayedParcels())
	totalErrors := 0

	var err error

	for {
		select {
		case <-n.ctx.Done():
			return nil, n.ctx.Err()
		case <-ctx.Done():
			return nil, ctx.Err()
		case r := <-responseChan:
			if tmperr := n.validateRrpcResponse(r); tmperr != nil {
				totalErrors += 1

				if totalErrors > n.MaxErrors() {
					err = tmperr

					return nil, err
				}

				continue
			}

			responses = append(responses, r.Response)

			if len(responses) >= n.MinimalRelayedParcels() {
				return responses, nil
			}
		}

	}
}

func (n *coordinator) validateRrpcResponse(r relayResponse) error {
	switch {
	case r.RelayStreamError != nil:
		return status.ErrorProto(r.RelayStreamError)

	case r.Response == nil:
		return status.Error(codes.Internal, "empty response")

	case r.Response.RpcError != nil:
		return status.ErrorProto(r.Response.RpcError)

	default:
		return nil
	}

}

func (n *coordinator) CancelRequest(uuid string) {
	for _, conn := range n.relayConns {
		conn.cancelRequest(uuid)
	}
}

func (n *coordinator) PublishProof(p *Proof) {
	for _, conn := range n.relayConns {
		conn.SendProof(p)
	}
}

func (n *coordinator) GetRelayConn(hostname string) *RelayConn {
	return n.relayConns[hostname]
}

// NewCoordinator creates a Coordinator that is tied to a speicific node. cannot reuse for different nodes on same machine!
func NewCoordinator(netdata NetworkData, skey crypto.PrivateKey) ServerCoordinator {
	maxErasures := netdata.MaxErrors()
	numDataShards := len(netdata.Servers()) - maxErasures

	ecc_, err := ecc.NewRSEncoderDecoder(numDataShards, maxErasures)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	v := NewVerifier(runtime.NumCPU())
	myAddress := netdata.GetHostname(skey.Public())
	ctx = AddIPToContext(ctx, myAddress)

	return &coordinator{
		NetworkData: netdata,
		skey:        skey,

		relayConns:  make(map[string]*RelayConn, len(netdata.Servers())),
		serverConns: make(map[string]*ServerConn, len(netdata.Servers())),

		MerkleCertVerifier: v,
		ecc:                ecc_,

		ctx:    ctx,
		cancel: cancel,

		myAddress: myAddress,

		callResponseChan: make(chan *CallStreamResponse, 100),
	}
}

func (n *coordinator) Close() error {
	defer n.MerkleCertVerifier.Stop()
	defer n.ecc.Stop()

	close(n.callResponseChan)

	n.cancel()

	var err error
	for _, conn := range n.relayConns {
		if err = conn.Close(); err != nil {
			fmt.Println("error closing relay connection:", err)
		}
	}

	for _, conn := range n.serverConns {
		if err = conn.Close(); err != nil {
			fmt.Println("error closing server connection:", err)
		}
	}

	return err
}

func (n *coordinator) Dial() error {
	for index, s := range n.NetworkData.Servers() {
		relayConn, err := NewRelayConn(n.ctx, s, index)
		if err != nil {
			return n.Close()
		}

		n.relayConns[s] = relayConn

		serverConn, err := newServerConn(n.ctx, s, n.callResponseChan)
		if err != nil {
			return n.Close()
		}
		n.serverConns[s] = serverConn
	}

	return nil
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

func (n *network) ContainKey(key []byte) bool {
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

// build new networkConfig type
func NewNetData(config *NetworkConfig) *network {
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

func (n *network) addServer(scnf ServerData) {
	n.serverAddresses = append(n.serverAddresses, scnf.Address)
	n.hostToPK[scnf.Address] = scnf.Publickey
	n.pkToHost[string(scnf.Publickey)] = scnf.Address
}

func isvalid(scnf ServerData) bool {
	return scnf.Address != "" && scnf.Publickey != nil
}
