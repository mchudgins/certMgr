package backend

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"text/template"

	log "github.com/Sirupsen/logrus"
	"github.com/mwitkow/go-grpc-middleware"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/context"

	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/mchudgins/certMgr/pkg/certMgr"
	"github.com/mchudgins/certMgr/pkg/healthz"
	pb "github.com/mchudgins/certMgr/pkg/service"
	"google.golang.org/grpc"
)

type server struct {
	cfg certMgr.AppConfig
	ca  *ca
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

// Run the backend command
func Run(cfg *certMgr.AppConfig) {

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}

	// create the Certificate Authority
	ca, err := NewCertificateAuthorityFromConfig(cfg)
	_ = ca

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
				grpcEndpointLog("certMgr")))
		pb.RegisterCertMgrServer(s, &server{cfg: *cfg})
		log.Infof("gRPC service listening on %s", cfg.GRPCListenAddress)
		errc <- s.Serve(lis)
	}()

	// http server
	go func() {
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
				log.WithError(err).WithField("template", "/").Errorf("Unable to parse template")
				return
			}

			err = tmp.Execute(w, data{Hostname: hostname})
			if err != nil {
				log.WithError(err).Error("Unable to execute template")
			}
		})

		log.Infof("HTTP service listening on %s", cfg.HTTPListenAddress)
		errc <- http.ListenAndServe(cfg.HTTPListenAddress, nil)
	}()

	// wait for somthin'
	log.Printf("exit: %s", <-errc)
}
