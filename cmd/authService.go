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
	"log"

	"github.com/spf13/cobra"
)

// authServiceCmd represents the authService command
var authServiceCmd = &cobra.Command{
	Use:   "authService",
	Short: "A trivial authentication service",
	Long: `The 'authService' is for development purposes only!

It responds 'true' to any token presented to it and sets the
the 'userID' to the value of the token supplied.

This means you can 'curl -H "Authorization: bearer IAmBob" http://myservice'`,

	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Work your own magic here
		log.Printf("'authService' started!  This command is for Development mode ONLY!")
	},
}

func init() {
	RootCmd.AddCommand(authServiceCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// authServiceCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// authServiceCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
