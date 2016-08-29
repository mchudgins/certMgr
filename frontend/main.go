package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/context"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/mchudgins/go-service-helper/pkg/loggingWriter"
	"github.com/mchudgins/go-service-helper/pkg/serveSwagger"
	"github.com/mchudgins/golang-service-starter/healthz"
	pb "github.com/mchudgins/golang-service-starter/service"
	"github.com/mchudgins/golang-service-starter/utils"
	"google.golang.org/grpc"
)

type server struct{}

var (
	swagger = MustAsset("../service/service.swagger.json")
	// boilerplate variables for good SDLC hygiene.  These are auto-magically
	// injected by the Makefile & linker working together.
	version   string
	buildTime string
	builder   string
	buildNum  string
	goversion string
)

func serveSwaggerData(w http.ResponseWriter, r *http.Request) {
	if !strings.HasSuffix(r.URL.Path, ".swagger.json") {
		log.Printf("Not Found: %s", r.URL.Path)
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(swagger)
}

// allowCORS allows Cross Origin Resource Sharing from any origin.
// Don't do this without consideration in production systems.
func allowCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			log.Printf("Origin: %s", origin)
			w.Header().Set("Access-Control-Allow-Origin", origin)
			if r.Method == "OPTIONS" && r.Header.Get("Access-Control-Request-Method") != "" {
				preflightHandler(w, r)
				return
			}
		}
		h.ServeHTTP(w, r)
	})
}

func preflightHandler(w http.ResponseWriter, r *http.Request) {
	headers := []string{"Content-Type", "Accept"}
	w.Header().Set("Access-Control-Allow-Headers", strings.Join(headers, ","))
	methods := []string{"GET", "HEAD", "POST", "PUT", "DELETE"}
	w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ","))
	log.Printf("preflight request for %s", r.URL.Path)
	return
}

func main() {
	log.Printf("golang-frontend-starter: version %s; buildTime: %s; built by: %s; buildNum: %s; (%s)",
		version, buildTime, builder, buildNum, goversion)

	cfg, err := utils.NewAppConfig()
	if err != nil {
		log.Printf("Unable to initialize the application (%s).  Exiting now.", err)
	}

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Starting app on host %s...", hostname)

	// make a channel to listen on events,
	// then launch the servers.

	errc := make(chan error)

	// interrupt handler
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	// http server
	go func() {
		ctx := context.Background()
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		mux := http.NewServeMux()
		gw := runtime.NewServeMux()

		hc, err := healthz.NewConfig(cfg)
		healthzHandler, err := healthz.Handler(hc)
		if err != nil {
			log.Panic(err)
		}

		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, `
{
"endpoints" :
	[
	"/api/v1/echo",
	"/healthz",
	"/metrics",
	"/swagger/",
	"/swagger-ui/"
	]
}
`)
		})

		mux.Handle("/api/v1/", gw)
		mux.Handle("/v1/", gw)
		mux.Handle("/healthz", healthzHandler)
		mux.Handle("/metrics", prometheus.Handler())
		mux.HandleFunc("/swagger/", serveSwaggerData)

		swaggerProxy, _ := serveSwagger.NewSwaggerProxy("/swagger-ui/")
		mux.HandleFunc("/swagger-ui/", swaggerProxy.ServeHTTP)

		/*
			http.HandleFunc("/v1/", func(w http.ResponseWriter, r *http.Request) {
				log.Printf("/api+")
				defer log.Printf("/api-")

				gw.ServeHTTP(w, r)
			})
		*/
		opts := []grpc.DialOption{grpc.WithInsecure()}
		err = pb.RegisterGreeterHandlerFromEndpoint(ctx, gw, cfg.GRPCListenAddress, opts)
		if err != nil {
			errc <- err
			return
		}

		log.Printf("HTTP service listening on %s", cfg.HTTPListenAddress)
		errc <- http.ListenAndServe(
			cfg.HTTPListenAddress,
			loggingWriter.HttpLogger(allowCORS(mux)))
	}()

	// wait for somthin'
	log.Printf("exit: %s", <-errc)
}
