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
	"github.com/mchudgins/certMgr/pkg/new-config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// newConfigFrontendCmd represents the frontend2 command
var newConfigFrontendCmd = &cobra.Command{
	Use:   "frontend <host.domain>",
	Short: "creates a new configuration file for the certMgr frontend (dev mode only)",
	Long: `This command creates a new configuration file for the frontend application.
For example:

	certmgr new-config frontend <host.domain> [flags]

An existing configuration can be updated with new certificates using
--config=<config file>; example:

	certmgr new-config frontend <host.docmain> --config=old-config.yaml

This command needs to be run in the 'certMgr' subdirectory with the CA files available.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := newConfigCmdConfig{}

		// some flags need special handling (sigh)
		cfg.Duration = viper.GetInt("duration")
		cfg.Verbose = viper.GetBool("verbose")
		cfg.Config = viper.GetString("config")

		new_config.RunFrontendConfig(cmd, args, cfg.Config, cfg.Duration, cfg.Verbose)
	},
}

func init() {
	newConfigCmd.AddCommand(newConfigFrontendCmd)
}
