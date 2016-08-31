package utils

import (
	"os"

	"github.com/spf13/cobra"
)

// AppConfig provides the global configuration of the application
type AppConfig struct {
	HTTPListenAddress string
	GRPCListenAddress string
}

// NewAppConfig sets up all the basic configuration data from flags, env, etc
func NewAppConfig(cmd *cobra.Command) (*AppConfig, error) {

	grpcAddr, err := cmd.Flags().GetString("grpc")
	if err != nil {
	}

	httpAddr, err := cmd.Flags().GetString("http")
	if err != nil {
		return nil, err
	}

	grpc := os.Getenv("GRPC_ADDRESS")
	if len(grpc) == 0 {
		grpc = grpcAddr
	}

	addr := os.Getenv("HTTP_ADDRESS")
	if len(addr) == 0 {
		addr = httpAddr
	}

	return &AppConfig{HTTPListenAddress: addr, GRPCListenAddress: grpc}, nil
}
