// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
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
	"github.com/mchudgins/certMgr/pkg/newConfig"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// new-configCmd represents the new-config command
var newConfigCmd = &cobra.Command{
	Use:   "new-config",
	Short: "create a new configuration file for this application (dev mode only)",
	Long: `This command creates a new configuration file for the application.
For example:

	certmgr new-config frontend | backend | auth

This command needs to be run in the 'certMgr' subdirectory with the CA files available.`,

	PersistentPreRun: func(cmd *cobra.Command, args []string) {

		// some flags need special handling (sigh)
		viper.BindPFlags(cmd.Flags())
		newConfig.NewConfigDefault.Duration = viper.GetInt("duration")
		newConfig.NewConfigDefault.Verbose = viper.GetBool("verbose")
		newConfig.NewConfigDefault.Config = viper.GetString("config")
		newConfig.NewConfigDefault.WriteKey = viper.GetBool("write-key")
	},

	Run: func(cmd *cobra.Command, args []string) {
		cmd.Usage()
	},
}

func init() {
	RootCmd.AddCommand(newConfigCmd)
	newConfigCmd.PersistentFlags().Int("duration", newConfig.NewConfigDefault.Duration, "# of days duration for the certificate's validity")
	newConfigCmd.PersistentFlags().Bool("write-key", newConfig.NewConfigDefault.WriteKey, "write out the key for the generated cert")
}
