package rrpc

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MethodHandler func(server interface{}, ctx context.Context, dec func(interface{}) error) (interface{}, error)

type MethodDesc struct {
	Name    string
	Handler MethodHandler
}
type ServiceDesc struct {
	ServerType        interface{} // the type of the service implementation
	Server            interface{} // the Server for service methods
	MethodDescriptors map[string]*MethodDesc
}

type Services map[string]*ServiceDesc

func extractMethodInfo(name string) (*struct{ serviceName, methodName string }, error) {
	if name != "" && name[0] == '/' {
		name = name[1:]
	}

	pos := strings.LastIndex(name, "/")
	if pos == -1 {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("method name %v could not be parsed", name))
	}

	retVal := struct{ serviceName, methodName string }{name[:pos], name[pos+1:]}

	return &retVal, nil
}

func (s Services) getServiceAndMethodDesc(name string) (*ServiceDesc, *MethodDesc, error) {
	info, err := extractMethodInfo(name)
	if err != nil {
		return nil, nil, err
	}

	service, svsExist := s[info.serviceName]
	if !svsExist {
		return nil, nil, status.Error(codes.InvalidArgument, fmt.Sprintf("service %v does not exist", info.serviceName))
	}

	methodDesc, methodExist := service.MethodDescriptors[info.methodName]
	if !methodExist {
		return nil, nil, status.Error(
			codes.InvalidArgument,
			fmt.Sprintf("method %v in service %v does not exist", info.methodName, info.serviceName),
		)
	}

	return service, methodDesc, nil
}

func SingleServiceFromGRPCServiceDesc(dsc grpc.ServiceDesc, hndl interface{}) Services {
	mdscs := map[string]*MethodDesc{}
	for _, m := range dsc.Methods {
		m := m
		mdscs[m.MethodName] = &MethodDesc{
			Name: m.MethodName,
			Handler: func(server interface{}, ctx context.Context, dec func(interface{}) error) (interface{}, error) {
				return m.Handler(server, ctx, dec, nil)
			},
		}
	}

	return Services{
		dsc.ServiceName: {
			ServerType:        dsc.HandlerType,
			Server:            hndl,
			MethodDescriptors: mdscs,
		},
	}
}
