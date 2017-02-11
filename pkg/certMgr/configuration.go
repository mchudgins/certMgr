package certMgr

// AppConfig provides the global configuration of the application.
type AppConfig struct {
	Certificate        string `json:",omitempty"` // the pem-encoded certificate for the service
	CertFilename       string `json:",omitempty"` // the name of the file containing the pem-encoded certicate for the service
	Insecure           bool   // for testing purposes, do not start-up TLS endpoints
	KeyFilename        string // the name of the file containing the pem-encoded key for the service's cert
	Config             string `json:",omitempty"` // load config data from this file (may be a url)
	HTTPListenAddress  string
	GRPCListenAddress  string
	AuthServiceAddress string
	Verbose            bool

	// specific config options for each command & subcommand
	Backend BackendConfig
}

type BackendConfig struct {
	AuthorizedCreators   []string // users authorized to create new certificates (an empty list permits anyone)
	Bundle               string   // the pem-encoded bundle of intermediate CA's
	SigningCACertificate string   `json:",omitempty"` // the pem-encoded signing CA
	SigningCAKeyFilename string   `json:",omitempty"` // filename for the CA key
	MaxDuration          int      // maximum # of days this CA will issue a cert
}

// the default configuration
var (
	DefaultAppConfig = &AppConfig{
		CertFilename:       "cert.pem",
		KeyFilename:        "key.pem",
		Config:             "",
		HTTPListenAddress:  ":8443",
		GRPCListenAddress:  ":50051",
		AuthServiceAddress: "authn:50051",
		Insecure:           false,
		Verbose:            false,
		Backend:            defaultBackendConfig,
	}

	// defaultConfig holds default values
	defaultBackendConfig = BackendConfig{
		AuthorizedCreators:   []string{""},
		SigningCAKeyFilename: "ca-key.pem",
		MaxDuration:          365, // max duration, in days, for any certificate
	}
)
