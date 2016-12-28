// Copyright © 2016 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/mchudgins/certMgr/pkg/certMgr"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "certMgr",
	Short: "Provides an easy-to-use means for managing self-signed certificates",
	Long: `certMgr provides frontend and backend services for managing self-signed
certificates via the inter-webs.  It also provides a CLI for boot-strapping
certMgr itself.  Examples:

	certMgr backend --http :8080
		starts the backend service listening on port 8080.

	certMgr frontend
		starts the frontend service.

	certMgr new certMgr.example.com
		creates a new certificate and key.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports Persistent Flags, which, if defined here,
	// will be global for your application.

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.certMgr.yaml)")
	RootCmd.PersistentFlags().String("grpc", certMgr.DefaultAppConfig.GRPCListenAddress, "listen address for the gRPC server")
	RootCmd.PersistentFlags().String("http", certMgr.DefaultAppConfig.HTTPListenAddress, "listen address for the http server")
	RootCmd.PersistentFlags().String("auth", certMgr.DefaultAppConfig.AuthServiceAddress, "gRPC port for Auth Service")
	RootCmd.PersistentFlags().BoolP("verbose", "v", false, "provide verbose output")
	RootCmd.PersistentFlags().String("certFilename", certMgr.DefaultAppConfig.CertFilename,
		"filename of the pem-encoded certificate for the service")
	RootCmd.PersistentFlags().String("certificate", certMgr.DefaultAppConfig.Certificate,
		"the pem-encoded certificate for the service")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	log.SetLevel(log.InfoLevel)

	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName(".certMgr") // name of config file (without extension)
	viper.AddConfigPath("$HOME")    // adding home directory as first search path
	viper.AutomaticEnv()            // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		// fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
