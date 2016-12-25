package certMgr

// AppConfig provides the global configuration of the application.
type AppConfig struct {
	Config             string `json:"config"`
	HTTPListenAddress  string `json:"http"`
	GRPCListenAddress  string `json:"grpc"`
	AuthServiceAddress string `json:"auth"`
	Verbose            bool   `json:"verbose"`

	// specific config options for each command & subcommand
	Backend BackendConfig
}

type BackendConfig struct {
	Bundle               string // the pem-encoded bundle of intermediate CA's
	KeyFilename          string // the name of the file containing the pem-encoded key for the signing CA
	SigningCACertificate string // the pem-encoded signing CA
	SigningCAKeyFilename string // filename for the CA key
	MaxDuration          int    // maximum # of days this CA will issue a cert
}

// the default configuration
var (
	DefaultAppConfig = &AppConfig{
		Config:             "",
		HTTPListenAddress:  ":8080",
		GRPCListenAddress:  ":50051",
		AuthServiceAddress: "auth.dstcorp.net:443",
		Verbose:            false,
		Backend:            defaultBackendConfig,
	}

	// defaultConfig holds default values
	defaultBackendConfig = BackendConfig{
		SigningCAKeyFilename: "ca-key.pem",
		KeyFilename:          "key.pem",
		MaxDuration:          365, // max duration, in days, for any certificate
	}
)
