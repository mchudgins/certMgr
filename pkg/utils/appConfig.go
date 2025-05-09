package utils

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"path"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/afex/hystrix-go/hystrix"
	"github.com/mchudgins/certMgr/pkg/certMgr"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// readConfigFile
func readConfigFile(uri string) error {

	// what kind of config file?
	ext := path.Ext(uri)
	viper.SetConfigType(ext[1:])

	// did they include a "file://"?
	filename := uri
	if strings.HasPrefix(uri, "file://") {
		filename = uri[0:len("file://")]
	}

	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return viper.MergeConfig(file)
}

// readConfigViaNet
func readConfigViaNet(uri string) error {

	// what kind of config file?
	ext := path.Ext(uri)
	viper.SetConfigType(ext[1:])

	c := &http.Client{}
	err := hystrix.Do("configServer", func() (err error) {
		resp, err := c.Get(uri)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		return viper.MergeConfig(resp.Body)
	}, nil)

	return err
}

// readConfig
func readConfig(uri string) error {

	switch uri[0:5] {
	case "http:":
		return readConfigViaNet(uri)

	case "file:":
		return readConfigFile(uri[7:])

	default:
		log.Printf("Warning: unable to interpret %s as a file or network location.", uri)
	}

	return nil
}

func NewConfig(cmd *cobra.Command, defaultConfig interface{}, cfg interface{}) error {
	defaultSettings, err := json.Marshal(defaultConfig)
	if err != nil {
		panic(err)
	}

	viper.SetConfigType("json")
	viper.ReadConfig(bytes.NewReader(defaultSettings))
	viper.SetEnvPrefix("certMgr")
	viper.BindPFlags(cmd.Flags())

	// afore we do anything, get the value for the config server,
	// download the config, and feed it into viper
	configURI := viper.GetString("config")

	if len(configURI) == 0 {
		// fetch the config from the default location, $HOME/.certMgr.yaml
		home := os.Getenv("HOME")
		configURI = "file://" + home + "/.certMgr.yaml"
	}

	err = readConfig(configURI)
	if err != nil {
		log.Printf("Warning: unable to obtain configuration from %s -- %s", configURI, err)
	}

	err = viper.Unmarshal(cfg)
	if err != nil {
		panic(err)
	}

	return nil
}

// NewAppConfig sets up all the basic configuration data from flags, env, etc
func NewAppConfig(cmd *cobra.Command) (*certMgr.AppConfig, error) {
	var activeConfig certMgr.AppConfig

	err := NewConfig(cmd, certMgr.DefaultAppConfig, &activeConfig)
	if err != nil {
		panic(err)
	}

	// these flags need special handling 'cause
	// the flag name and the field name don't match (sigh)
	activeConfig.HTTPListenAddress = viper.GetString("http")
	activeConfig.GRPCListenAddress = viper.GetString("grpc")
	activeConfig.AuthServiceAddress = viper.GetString("auth")

	//	log.Printf("Current config:  %+v", activeConfig)
	return &activeConfig, nil
}
