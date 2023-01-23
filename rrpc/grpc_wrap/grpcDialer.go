package rrpcnet

import (
	"github.com/jonathanMweiss/resmix/internal/codec"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
)

type CustomCodec struct{}

func (c CustomCodec) String() string {
	return c.Name()
}

func init() {
	encoding.RegisterCodec(CustomCodec{})

}
func (c CustomCodec) Marshal(v interface{}) ([]byte, error) {
	return codec.Marshal(v)
	//if err != nil {
	//	panic(err)
	//}
	//return a, nil
}

func (c CustomCodec) Unmarshal(data []byte, v interface{}) error {
	return codec.Unmarshal(data, v)
	//if err != nil {
	//	err := codec.Unmarshal(data, v)
	//	panic(err)
	//}
	//return nil
}

func (c CustomCodec) Name() string {
	return "rrpc"
}

func GetGrpcClient(target string, grpcOpts ...grpc.DialOption) (*grpc.ClientConn, error) {
	// Now i can edit every connection at once
	opts := append(grpcOpts, grpc.WithDefaultCallOptions(grpc.CallContentSubtype(CustomCodec{}.String())))
	return grpc.Dial(target, opts...)
}
