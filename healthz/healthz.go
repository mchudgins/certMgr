// modeled after github.com/kelseyhightower/app-healthz2
// so go look there for additional ideas related to health checking:
// databases, vault, etc.

package healthz

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/mchudgins/golang-service-starter/utils"
)

// Config provides data for the healthz handler
type Config struct {
	Hostname string
	//	Database DatabaseConfig
	//	Vault    VaultConfig
}

type handler struct {
	// dc       *DatabaseChecker
	// vc       *VaultChecker
	hostname string
	metadata map[string]string
}

// NewConfig initializes a healthz.Config struct
func NewConfig(appConfig *utils.AppConfig) (*Config, error) {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}

	hc := &Config{
		Hostname: hostname,
	}

	return hc, nil
}

// Handler provides a new healthz handler
func Handler(hc *Config) (http.Handler, error) {
	metadata := make(map[string]string)

	h := &handler{hc.Hostname, metadata}
	return h, nil
}

type Response struct {
	Hostname string            `json:"hostname"`
	Metadata map[string]string `json:"metadata"`
	Errors   []Error           `json:"errors"`
}

type Error struct {
	Description string            `json:"description"`
	Error       string            `json:"error"`
	Metadata    map[string]string `json:"metadata"`
	Type        string            `json:"type"`
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	response := Response{
		Hostname: h.hostname,
		Metadata: h.metadata,
	}

	statusCode := http.StatusOK

	errors := make([]Error, 0)

	response.Errors = errors
	if len(response.Errors) > 0 {
		statusCode = http.StatusInternalServerError
		for _, e := range response.Errors {
			log.Println(e.Error)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	data, err := json.MarshalIndent(&response, "", "  ")
	if err != nil {
		log.Println(err)
	}
	w.Write(data)
}
