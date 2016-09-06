package utils

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"path"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// AppConfig provides the global configuration of the application.
type AppConfig struct {
	HTTPListenAddress string `json:"http"`
	GRPCListenAddress string `json:"grpc"`
	User              User   `json:"user"`
}

type User struct {
	UserID string `json:"userID"`
	Name   string `json:"name"`
}

type TestConfig struct {
	HTTPListenAddress string `json:"http"`
	User              User   `json:"user"`
}

var (
	defaultConfig = &AppConfig{
		HTTPListenAddress: ":8080",
		GRPCListenAddress: ":50051",
		User: User{
			UserID: "bones177",
			Name:   "Bob Jones",
		},
	}
)

// NewAppConfig sets up all the basic configuration data from flags, env, etc
func NewAppConfig(cmd *cobra.Command) (*AppConfig, error) {

	defaultSettings, err := json.Marshal(defaultConfig)
	if err != nil {
		panic(err)
	}
	log.Printf("defaultConfig    : %+v", defaultConfig)
	log.Printf("defaultSettings  : %s", defaultSettings)

	viper.SetConfigType("json")
	viper.ReadConfig(bytes.NewReader(defaultSettings))
	viper.SetEnvPrefix("fubar")
	log.Printf("user.name        : %s", viper.GetString("user.name"))
	log.Printf("FUBAR_GORF       : %s", viper.GetString("GORF"))
	log.Printf("gorf             : %s", viper.GetString("gorf"))

	viper.BindPFlags(cmd.Flags())
	log.Printf("httpaddress      : %s", viper.GetString("http"))

	log.Printf("GRPCListenAddress: %s", viper.GetString("GRPCListenAddress"))
	log.Printf("grpc             : %s", viper.GetString("grpc"))

	var activeConfig AppConfig
	err = viper.Unmarshal(&activeConfig)
	if err != nil {
		panic(err)
	}
	log.Printf("TestConfig.HTTPListenAddress: %s", activeConfig.HTTPListenAddress)
	log.Printf("TestConfig.GRPCListenAddress: %s", activeConfig.GRPCListenAddress)
	log.Printf("TestConfig.user.name        : %s", activeConfig.User.Name)
	log.Printf("TestConfig.user.userID      : %s", activeConfig.User.UserID)

	log.Printf("basename: %s", path.Base(os.Args[0]))

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
