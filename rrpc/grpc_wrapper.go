package rrpc

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type grpcServer struct {
	*grpc.Server
}

func newGrpcServer(desc Services, opts ...grpc.ServerOption) *grpcServer {
	gsrvr := grpc.NewServer(opts...)
	for serviceName, serviceDesc := range desc {
		i := 0
		ms := make([]grpc.MethodDesc, len(serviceDesc.methodDescriptors))
		for nm, methodDesc := range serviceDesc.methodDescriptors {
			//for i := 0; i < len(serviceDesc.methodDescriptors); i++ {
			mdesc := methodDesc
			f := func(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
				// todo find a better way to do this. I don't want to add another function call.
				return mdesc.Handler(srv, ctx, dec)
			}

			ms[i] = grpc.MethodDesc{
				MethodName: nm,
				Handler:    f,
			}
			i += 1
		}

		s := grpc.ServiceDesc{
			ServiceName: serviceName,
			HandlerType: serviceDesc.serverType,
			Methods:     ms,
			Streams:     nil,
			Metadata:    "",
		}

		gsrvr.RegisterService(&s, serviceDesc.server)
	}

	return &grpcServer{gsrvr}
}

type grpcClient struct {
	conn *grpc.ClientConn
}

func (g *grpcClient) Close() error {
	return g.conn.Close()
}

var bypassDoesNotImplement = status.Error(codes.Unimplemented, "grpc doesn't have this functionality")

func (g *grpcClient) DirectCall(rq *Request) error {
	if g.conn == nil {
		return status.Error(codes.Internal, "nil client")
	}
	ctx := rq.Context
	return g.conn.Invoke(ctx, rq.Method, rq.Args, rq.Reply)
}

func (g *grpcClient) AsyncDirectCall(rq *Request) (<-chan error, error) {
	errchn := make(chan error, 1)

	go func() {
		if g.conn == nil {
			errchn <- status.Error(codes.Internal, "nil client")

			return
		}

		errchn <- g.conn.Invoke(rq.Context, rq.Method, rq.Args, rq.Reply)
	}()

	return errchn, nil
}

func (g *grpcClient) RobustCall(request *Request) error {
	return bypassDoesNotImplement
}

func newGrpcClient(target string, opts ...grpc.DialOption) (ClientConn, error) {
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	c, err := grpc.Dial(target, opts...)
	if err != nil {
		return nil, err

	}

	return &grpcClient{conn: c}, nil
}
