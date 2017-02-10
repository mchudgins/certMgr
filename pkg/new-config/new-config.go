package new_config

import (
	"context"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/ghodss/yaml"
	"github.com/mchudgins/certMgr/pkg/backend"
	"github.com/mchudgins/certMgr/pkg/certMgr"
	"github.com/spf13/cobra"
)

func RunBackendConfig(cmd *cobra.Command, args []string, config string, duration int, fVerbose bool) {

	if fVerbose {
		log.SetLevel(log.DebugLevel)
	}

	if len(args) != 1 {
		cmd.Usage()
		os.Exit(1)
	}

	serverName := args[0]

	var cfg *certMgr.AppConfig
	if config == nil || len(config) == 0 {
		cfg = certMgr.DefaultAppConfig
	} else {

	}

	// generate the requested certificate
	// initialize the SimpleCA
	ca, err := backend.NewCertificateAuthority("signingCert",
		"ca/cap/cap-ca.crt",
		"ca/cap/private/cap-ca.key",
		"ca/cap/cap-ca.crt")
	if err != nil {
		log.WithError(err).Fatal("unable to initialize the CA")
	}

	ctx := context.Background()
	var empty []string
	cert, _, err := ca.CreateCertificate(ctx, serverName, empty, time.Duration(duration)*time.Hour*24)
	if err != nil {
		log.WithError(err).
			WithField("name", serverName).
			Fatal("unable to create certificate")
		os.Exit(3)
	}

	cfg.Certificate = cert

	y, err := yaml.Marshal(cfg)
	if err != nil {
		log.WithError(err).
			Fatal("while marshaling configuration to yaml")
		os.Exit(1)
	}

	os.Stdout.Write(y)
}
