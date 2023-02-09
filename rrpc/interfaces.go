package rrpc

import (
	"fmt"
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
	ServerCoordinator ServerCoordinator
}

type ServerConfigs struct {
	CommonConfigs
	GrpcOptions []grpc.ServerOption
}

func (c CommonConfigs) isEmpty() bool {
	return c.ServerCoordinator == nil && !c.IsGrpc
}

type ClientConfigs struct {
	CommonConfigs
	GrpcOptions []grpc.DialOption
}

var ErrNoConfigs = fmt.Errorf("no configs provided")

func NewServer(services Services, cnfgs ServerConfigs) (Server, error) {
	if cnfgs.isEmpty() {
		return nil, ErrNoConfigs
	}

	if cnfgs.IsGrpc {
		return newGrpcServer(services, cnfgs.GrpcOptions...), nil
	}

	return newServerService(
		services,
		cnfgs.ServerCoordinator,
		cnfgs.GrpcOptions...,
	)
}

func NewConnection(target string, cnfgs ClientConfigs) (ClientConn, error) {
	if cnfgs.isEmpty() {
		return nil, ErrNoConfigs
	}

	if cnfgs.IsGrpc {
		return newGrpcClient(target, cnfgs.GrpcOptions...)
	}

	return newClient(target, cnfgs.ServerCoordinator), nil

}
