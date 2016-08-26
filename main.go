package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"text/template"

	"github.com/mwitkow/go-grpc-middleware"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/context"

	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/mchudgins/golang-backend-starter/healthz"
	"github.com/mchudgins/golang-backend-starter/utils"
	"google.golang.org/grpc"
)

type server struct{}

var (
	// boilerplate variables for good SDLC hygiene.  These are auto-magically
	// injected by the Makefile & linker working together.
	version   string
	buildTime string
	builder   string
	goversion string
)

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *HelloRequest) (*HelloReply, error) {
	return &HelloReply{Message: "Hello " + in.Name}, nil
}

func grpcEndpointLog(s string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (interface{}, error) {
		log.Printf("grpcEndpointLog %s+", s)
		defer log.Printf("grpcEndpointLog %s-", s)
		return handler(ctx, req)
	}
}

func main() {
	cfg, err := utils.NewAppConfig()
	if err != nil {
		log.Printf("Unable to initialize the application (%s).  Exiting now.", err)
	}

	log.Printf("Starting app...")

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}

	hc, err := healthz.NewConfig(cfg)
	healthzHandler, err := healthz.Handler(hc)
	if err != nil {
		log.Panic(err)
	}

	http.Handle("/healthz", healthzHandler)
	http.Handle("/metrics", prometheus.Handler())
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

	// make a channel to listen on events,
	// then launch the servers.

	errc := make(chan error)

	// interrupt handler
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	// gRPC server
	go func() {
		lis, err := net.Listen("tcp", cfg.GRPCListenAddress)
		if err != nil {
			errc <- err
			return
		}

		s := grpc.NewServer(
			grpc_middleware.WithUnaryServerChain(
				grpc_prometheus.UnaryServerInterceptor,
				grpcEndpointLog("hello")))
		RegisterGreeterServer(s, &server{})
		log.Printf("gRPC service listening on %s", cfg.GRPCListenAddress)
		errc <- s.Serve(lis)
	}()

	// http server
	go func() {
		log.Printf("HTTP service listening on %s", cfg.HTTPListenAddress)
		errc <- http.ListenAndServe(cfg.HTTPListenAddress, nil)
	}()

	// wait for somthin'
	log.Printf("exit: %s", <-errc)
}
