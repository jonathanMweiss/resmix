package rrpcnet

import "google.golang.org/grpc"

func GetGrpcServer(opts ...grpc.ServerOption) *grpc.Server {
	// grpc for some reason does not search for the wanted codec, so im forcing it.
	return grpc.NewServer(opts...)
}
