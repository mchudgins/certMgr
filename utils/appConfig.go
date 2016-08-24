package utils

import (
	"flag"
	"os"
)

// AppConfig provides the global configuration of the application
type AppConfig struct {
	HTTPListenAddress string
	GRPCListenAddress string
}

var (
	listenAddr = flag.String("http", ":8080", "listen address for the http server")
	grpcAddr   = flag.String("grpc", ":50051", "listen address for the gRPC server")
)

// NewAppConfig sets up all the basic configuration data from flags, env, etc
func NewAppConfig() (*AppConfig, error) {

	flag.Parse()

	addr := os.Getenv("HTTP_ADDRESS")
	if len(addr) == 0 {
		addr = *listenAddr
	}

	grpc := os.Getenv("GRPC_ADDRESS")
	if len(grpc) == 0 {
		grpc = *grpcAddr
	}

	return &AppConfig{HTTPListenAddress: addr, GRPCListenAddress: grpc}, nil
}
