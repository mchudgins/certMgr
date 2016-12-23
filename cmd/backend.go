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
	log "github.com/Sirupsen/logrus"
	"github.com/mchudgins/certMgr/pkg/backend"
	"github.com/mchudgins/certMgr/pkg/certMgr"
	"github.com/mchudgins/certMgr/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// backendCmd represents the backend command
var backendCmd = &cobra.Command{
	Use:   "backend",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		// determine the backend app's configuration
		cfg, err := utils.NewAppConfig(cmd)
		if err != nil {
			log.WithField("error", err).Fatal("an error occurred while obtaining the application configuration")
		}

		// these flags must be handled individually since the flag name doesn't match the field name
		// they are in a sub-struct of the top level structure
		cfg.Backend.KeyFilename = viper.GetString("key")

		// set the log level
		if cfg.Verbose {
			log.SetLevel(log.DebugLevel)
		}

		utils.StartUpMessage()

		log.Debugf("Current config:  %+v", cfg)

		// ready to run...
		backend.Run(cfg)
	},
}

func init() {
	RootCmd.AddCommand(backendCmd)

	backendCmd.PersistentFlags().String("key", certMgr.DefaultAppConfig.Backend.KeyFilename, "key filename")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// backendCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// backendCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
