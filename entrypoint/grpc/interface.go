package entrypointGrpc

import "google.golang.org/grpc"

type GrpcService interface {
	Register(server *grpc.Server)
}

type Config interface {
	ServiceGrpcAddress() string
}
