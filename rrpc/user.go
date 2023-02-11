package rrpc

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func insertUuidToContext(ctx context.Context, uuid string) context.Context {
	return context.WithValue(ctx, "uuid", uuid)
}

func ExtractUuidFromContext(ctx context.Context) (string, error) {
	uuid, ok := ctx.Value("uuid").(string)
	if !ok {
		return "", status.Error(codes.Internal, "couldn't extract uuid from context")
	}
	return uuid, nil
}
