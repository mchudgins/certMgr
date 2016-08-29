package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/context"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/mchudgins/golang-service-starter/healthz"
	"github.com/mchudgins/golang-service-starter/pkg/serveSwagger"
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

type loggingWriter struct {
	w             http.ResponseWriter
	statusCode    int
	contentLength int
}

func NewLoggingWriter(w http.ResponseWriter) *loggingWriter {
	return &loggingWriter{w: w}
}

func (l *loggingWriter) Header() http.Header {
	return l.w.Header()
}

func (l *loggingWriter) Write(data []byte) (int, error) {
	l.contentLength += len(data)
	return l.w.Write(data)
}

func (l *loggingWriter) WriteHeader(status int) {
	log.Printf("http status: %d", status)
	l.statusCode = status
	l.w.WriteHeader(status)
}

func (l *loggingWriter) Length() int {
	return l.contentLength
}

func (l *loggingWriter) StatusCode() int {
	return l.statusCode
}

// httpLogger provides per request log statements (ala Apache httpd)
func httpLogger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lw := NewLoggingWriter(w)
		defer func() {
			end := time.Now()
			duration := end.Sub(start)
			log.Printf("host: %s; uri: %s; remoteAddress: %s; method:  %s; proto: %s; status: %d, contentLength: %d, duration: %.3f; ua: %s",
				r.Host,
				r.RemoteAddr,
				r.URL,
				r.Method,
				r.Proto,
				lw.StatusCode(),
				lw.Length(),
				duration.Seconds()*1000,
				r.UserAgent())
		}()

		h.ServeHTTP(lw, r)

	})
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
	"/v1/echo",
	"/healthz",
	"/metrics",
	"/swagger/",
	"/swagger-ui/"
	]
}
`)
		})

		mux.Handle("/v1/", gw)
		mux.Handle("/healthz", healthzHandler)
		mux.Handle("/metrics", prometheus.Handler())
		mux.HandleFunc("/swagger/", serveSwaggerData)
		mux.HandleFunc("/swagger-ui/", serveSwagger.ServeHTTP)
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
		errc <- http.ListenAndServe(cfg.HTTPListenAddress, httpLogger(allowCORS(mux)))
	}()

	// wait for somthin'
	log.Printf("exit: %s", <-errc)
}
