package newConfig

import (
	"context"
	"os"
	"time"

	"io/ioutil"

	"io"

	log "github.com/Sirupsen/logrus"
	"github.com/ghodss/yaml"
	"github.com/mchudgins/certMgr/pkg/backend"
	"github.com/mchudgins/certMgr/pkg/certMgr"
	"github.com/mchudgins/certMgr/pkg/utils"
	"github.com/spf13/cobra"
)

type NewConfigCmdConfig struct {
	Duration int    `json:"duration"`
	Verbose  bool   `json:"verbose"`
	Config   string `json:"config"`
	WriteKey bool   `json:"writeKeyfile"`
}

// defaultConfig holds default values
var NewConfigDefault = &NewConfigCmdConfig{
	Duration: 90,
	Verbose:  false,
	WriteKey: false,
}

// generate the cert (optionally writing the the key to the keyfile) and return the updated cfg
func generateCert(cmd *cobra.Command, serverName string, bundle string, cmdConfig *NewConfigCmdConfig,
	cfg *certMgr.AppConfig) (*certMgr.AppConfig, error) {

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
	cert, key, err := ca.CreateCertificate(ctx, serverName, empty, time.Duration(cmdConfig.Duration)*time.Hour*24)
	if err != nil {
		log.WithError(err).
			WithField("name", serverName).
			Fatal("unable to create certificate")
	}

	cfg.Certificate = string(cert[:]) + bundle
	// zap the cert file name field since we include the actual cert in the config
	cfg.CertFilename = ""

	// write the key, if requested
	if cmdConfig.WriteKey {
		var keyfile io.Writer
		if len(cfg.KeyFilename) > 0 {
			file, err := os.OpenFile(cfg.KeyFilename, os.O_WRONLY|os.O_CREATE, 0600)
			if err != nil {
				log.WithError(err).
					WithField("filename", cfg.KeyFilename).
					Fatal("while opening keyfile")
			}
			defer file.Close()
			keyfile = file
		} else {
			keyfile = cmd.OutOrStdout()
		}
		keyfile.Write(key[:])
	}

	return cfg, nil
}

func RunAuthConfig(cmd *cobra.Command, args []string, cmdConfig *NewConfigCmdConfig) {
	RunBackendConfig(cmd, args, cmdConfig)
}

func RunFrontendConfig(cmd *cobra.Command, args []string, cmdConfig *NewConfigCmdConfig) {
	RunBackendConfig(cmd, args, cmdConfig)
}

func RunBackendConfig(cmd *cobra.Command, args []string, cmdConfig *NewConfigCmdConfig) {

	if cmdConfig.Verbose {
		log.Info("verbose mode ON")
		log.SetLevel(log.DebugLevel)
	}

	if len(args) != 1 {
		cmd.Usage()
		os.Exit(1)
	}

	serverName := args[0]

	var cfg *certMgr.AppConfig
	if len(cmdConfig.Config) == 0 {
		cfg = certMgr.DefaultAppConfig
	} else {
		log.WithField("config", cmdConfig.Config).Debug("")
		file, err := utils.OpenReadCloser(cmdConfig.Config)
		if err != nil {
			log.WithError(err).WithField("config", cmdConfig.Config).
				Fatal("exiting")
		}
		defer file.Close()

		buf, err := ioutil.ReadAll(file)
		if err != nil {
			log.WithError(err).WithField("config", cmdConfig.Config).
				Fatal("exiting")
		}

		cfg = &certMgr.AppConfig{}
		err = yaml.Unmarshal(buf, cfg)
		if err != nil {
			log.WithError(err).WithField("config", cmdConfig.Config).
				Fatal("exiting")
		}
	}

	// fetch the bundle
	bundleFile, err := utils.OpenReadCloser("file://ca/cap/ca-bundle.pem")
	if err != nil {
		log.WithError(err).
			Fatal("unable to open ca-bundle.pem")
	}
	defer bundleFile.Close()
	buf, err := ioutil.ReadAll(bundleFile)
	if err != nil {
		log.WithError(err).
			Fatal("unable to read ca-bundle.pem")
	}
	cfg.Backend.Bundle = string(buf)

	// fetch the signing cert
	caSignerFile, err := utils.OpenReadCloser("file://ca/cap/cap-ca.crt")
	if err != nil {
		log.WithError(err).
			Fatal("unable to open cap-ca.crt")
	}
	defer caSignerFile.Close()
	buf, err = ioutil.ReadAll(caSignerFile)
	if err != nil {
		log.WithError(err).
			Fatal("unable to read cap-ca.crt")
	}
	cfg.Backend.SigningCACertificate = string(buf)

	cfg, err = generateCert(cmd, serverName, cfg.Backend.Bundle, cmdConfig, cfg)
	if err != nil {
		log.WithError(err).
			Fatal("while generating cert")
	}

	// write out the new config
	y, err := yaml.Marshal(cfg)
	if err != nil {
		log.WithError(err).
			Fatal("while marshaling configuration to yaml")
	}

	os.Stdout.Write(y)
}
