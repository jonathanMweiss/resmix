package rrpc

import (
	"context"
	"fmt"
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
	server            interface{} // the server for service methods
	methodDescriptors map[string]*MethodDesc
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

	methodDesc, methodExist := service.methodDescriptors[info.methodName]
	if !methodExist {
		return nil, nil, status.Error(
			codes.InvalidArgument,
			fmt.Sprintf("method %v in service %v does not exist", info.methodName, info.serviceName),
		)
	}

	return service, methodDesc, nil
}
