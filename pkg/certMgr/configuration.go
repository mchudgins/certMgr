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
	KeyFilename string `json:"keyFilename"`
	MaxDuration int    `json:"maxDuration"`
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
		KeyFilename: "key.pem",
		MaxDuration: 365,
	}
)
