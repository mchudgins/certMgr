// Copyright © 2016 Mike Hudgins <mchudgins@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/mchudgins/certMgr/pkg/backend"
	"github.com/mchudgins/certMgr/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type newCmdConfig struct {
	Config       string `json:"config"`
	CertFilename string `json:"certFilename"`
	KeyFilename  string `json:"keyFilename"`
	Duration     int    `json:"duration"`
}

// defaultConfig holds default values
var defaultConfig = &newCmdConfig{
	Config:       "",
	CertFilename: "cert.pem",
	KeyFilename:  "key.pem",
	Duration:     90,
}

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new <common name>",
	Short: "Create a new certificate",
	Long:  `Creates a new certificate and key for the specified common name.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Fprintf(cmd.OutOrStderr(), "fatal: a common name for the certificate must be provided on the command line\n")
			cmd.Usage()
			os.Exit(1)
		}

		cfg := &newCmdConfig{}

		err := utils.NewConfig(cmd, defaultConfig, cfg)
		if err != nil {
			log.WithField("error", err).Fatal("an error occurred while obtaining the application configuration")
		}

		// flags need special handling (sigh)
		cfg.CertFilename = viper.GetString("cert")
		cfg.KeyFilename = viper.GetString("key")
		cfg.Duration = viper.GetInt("duration")

		if viper.GetBool("verbose") {
			log.SetLevel(log.DebugLevel)
		}
		log.Debugf("Current config:  %+v", cfg)

		// initialize the SimpleCA
		backend.SimpleCA, err = backend.NewCertificateAuthority("signingCert",
			"ca/cap/cap-ca.crt",
			"ca/cap/private/cap-ca.key",
			"ca/cap/cap-ca.crt")
		if err != nil {
			log.WithField("error", err).Fatal("unable to initialize the CA")
		}
	},
}

func init() {
	RootCmd.AddCommand(newCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// newCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// newCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	newCmd.Flags().String("cert", defaultConfig.CertFilename, "output file for the PEM encoded certificate")
	newCmd.Flags().Int("duration", defaultConfig.Duration, "# of days duration for the certificate's validity")
	newCmd.Flags().String("key", defaultConfig.KeyFilename, "output file for the PEM encoded key")
}
