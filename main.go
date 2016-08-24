package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/mchudgins/golang-backend-starter/healthz"
	"github.com/mchudgins/golang-backend-starter/utils"
)

var (
	// boilerplate variables for good SDLC hygiene.  These are auto-magically
	// injected by the Makefile & linker working together.
	version   string
	buildTime string
	builder   string
	goversion string
)

func main() {
	cfg, err := utils.NewAppConfig()
	if err != nil {
		log.Fatalf("Unable to initialize the application (%s).  Exiting now.", err)
	}

	log.Println("Starting app...")

	hc, err := healthz.NewConfig(cfg)
	healthzHandler, err := healthz.Handler(hc)
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/healthz", healthzHandler)

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		type data struct {
			Hostname string
		}

		tmp, err := template.New("/").Parse(html)
		if err != nil {
			fmt.Fprintf(w, "Unable to parse template: %s", err)
			return
		}

		err = tmp.Execute(w, data{Hostname: hostname})
		if err != nil {
			fmt.Fprintf(w, "Unable to execute template: %s", err)
		}
	})

	log.Printf("HTTP service listening on %s", cfg.HTTPListenAddress)
	err = http.ListenAndServe(cfg.HTTPListenAddress, nil)
	log.Printf("ListenAndServe:  %s", err)
}
