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

import "github.com/spf13/cobra"

type newConfigCmdConfig struct {
	Duration int    `json:"duration"`
	Verbose  bool   `json:"verbose"`
	Config   string `json:"config"`
}

// defaultConfig holds default values
var defaultNewConfig = &newConfigCmdConfig{
	Duration: 90,
	Verbose:  false,
}

// new-configCmd represents the new-config command
var newConfigCmd = &cobra.Command{
	Use:   "new-config",
	Short: "create a new configuration file for this application (dev mode only)",
	Long: `This command creates a new configuration file for the application.
For example:

	certmgr new-config frontend | backend | auth

This command needs to be run in the 'certMgr' subdirectory with the CA files available.`,

	Run: func(cmd *cobra.Command, args []string) {
		cmd.Usage()
	},
}

func init() {
	RootCmd.AddCommand(newConfigCmd)
	newConfigCmd.PersistentFlags().Int("duration", defaultNewConfig.Duration, "# of days duration for the certificate's validity")
}
