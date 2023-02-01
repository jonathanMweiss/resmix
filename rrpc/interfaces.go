package rrpc

import (
	"github.com/jonathanMweiss/resmix/internal/crypto"
	"google.golang.org/grpc"
	"net"
)

type ClientConn interface {
	DirectCall(rq *Request) error
	AsyncDirectCall(rq *Request) (<-chan error, error)
	RobustCall(rq *Request) error

	Close() error
}

type Server interface {
	Serve(lis net.Listener) error
	Stop()
}

type CommonConfigs struct {
	IsGrpc            bool
	SecretKey         crypto.PrivateKey
	ServerCoordinator ServerCoordinator
}

type ServerConfigs struct {
	CommonConfigs
	GrpcOptions []grpc.ServerOption
}
type ClientConfigs struct {
	CommonConfigs
	GrpcOptions []grpc.DialOption
}

func NewServer(services Services, cnfgs ServerConfigs) (Server, error) {
	if cnfgs.IsGrpc {
		return newGrpcServer(services, cnfgs.GrpcOptions...), nil
	}

	return newServerService(cnfgs.SecretKey, services, cnfgs.ServerCoordinator, cnfgs.GrpcOptions...)
}

func NewConnection(target string, cnfgs ClientConfigs) (ClientConn, error) {
	if cnfgs.IsGrpc {
		return newGrpcClient(target, cnfgs.GrpcOptions...)
	}

	return newClient(cnfgs.SecretKey, target, cnfgs.ServerCoordinator), nil

}
