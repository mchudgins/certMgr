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
	"github.com/mchudgins/certMgr/pkg/frontend"
	"github.com/mchudgins/certMgr/pkg/utils"
	"github.com/spf13/cobra"
)

// frontendCmd represents the frontend command
var (
	frontendCmd = &cobra.Command{
		Use:   "frontend",
		Short: "launch a 'frontend' for the sample service",
		Long: `Launch a 'frontend' API server for the sample service.

The frontend service accepts and provides JSON for the gRPC backend.`,
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := utils.NewAppConfig(cmd)
			if err != nil {
				log.Printf("Unable to initialize the application (%s).  Exiting now.", err)
			}

			utils.StartUpMessage(*cfg)

			frontend.Run(cfg)
		},
	}
)

func init() {
	RootCmd.AddCommand(frontendCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// frontendCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// frontendCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
