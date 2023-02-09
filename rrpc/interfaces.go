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

type Configs struct {
	IsGrpc            bool
	ServerCoordinator ServerCoordinator
	ServerOptions     []grpc.ServerOption
	DialOptions       []grpc.DialOption
}

func (c Configs) isEmpty() bool {
	return c.ServerCoordinator == nil && !c.IsGrpc
}

var ErrNoConfigs = fmt.Errorf("no configs provided")

func NewServer(services Services, cnfgs Configs) (Server, error) {
	if cnfgs.isEmpty() {
		return nil, ErrNoConfigs
	}

	if cnfgs.IsGrpc {
		return newGrpcServer(services, cnfgs.ServerOptions...), nil
	}

	return newServerService(
		services,
		cnfgs.ServerCoordinator,
		cnfgs.ServerOptions...,
	)
}

func NewConnection(target string, cnfgs Configs) (ClientConn, error) {
	if cnfgs.isEmpty() {
		return nil, ErrNoConfigs
	}

	if cnfgs.IsGrpc {
		return newGrpcClient(target, cnfgs.DialOptions...)
	}

	return newClient(target, cnfgs.ServerCoordinator), nil
}
