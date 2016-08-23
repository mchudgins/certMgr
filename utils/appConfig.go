package utils

import (
	"flag"
	"os"
)

// AppConfig provides the global configuration of the application
type AppConfig struct {
	HTTPListenAddress string
}

var (
	listenAddr = flag.String("listen", ":8080", "listen address for the reverse proxy")
)

// NewAppConfig sets up all the basic configuration data from flags, env, etc
func NewAppConfig() (*AppConfig, error) {

	flag.Parse()

	addr := os.Getenv("HTTP_ADDRESS")
	if len(addr) == 0 {
		addr = *listenAddr
	}

	return &AppConfig{HTTPListenAddress: addr}, nil
}
