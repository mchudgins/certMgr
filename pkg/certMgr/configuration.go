package certMgr

// AppConfig provides the global configuration of the application.
type AppConfig struct {
	CertFilename       string // the name of the file containing the pem-encoded certicate for the service
	KeyFilename        string // the name of the file containing the pem-encoded key for the service's cert
	Config             string // load config data from this file (may be a url)
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
	SigningCACertificate string   // the pem-encoded signing CA
	SigningCAKeyFilename string   // filename for the CA key
	MaxDuration          int      // maximum # of days this CA will issue a cert
}

// the default configuration
var (
	DefaultAppConfig = &AppConfig{
		CertFilename:       "cert.pem",
		KeyFilename:        "key.pem",
		Config:             "",
		HTTPListenAddress:  ":8080",
		GRPCListenAddress:  ":50051",
		AuthServiceAddress: "auth.dstcorp.net:443",
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
