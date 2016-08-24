package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"text/template"

	"golang.org/x/net/context"

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

	lis, err := net.Listen("tcp", cfg.GRPCListenAddress)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	RegisterGreeterServer(s, &server{})
	log.Printf("gRPC service listening on %s", cfg.GRPCListenAddress)
	go s.Serve(lis)

	log.Printf("HTTP service listening on %s", cfg.HTTPListenAddress)
	err = http.ListenAndServe(cfg.HTTPListenAddress, nil)
	log.Printf("ListenAndServe:  %s", err)
}
