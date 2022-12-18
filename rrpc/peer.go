package rrpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func GetPeerFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.InvalidArgument, "Missing caller identity")
	}
	ids, ok := md["caller"]
	if !ok {
		return "", status.Error(codes.InvalidArgument, "Missing caller identity")
	}
	return ids[0], nil
}

func AddIPToContext(ctx context.Context, addr string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "caller", addr)
}
