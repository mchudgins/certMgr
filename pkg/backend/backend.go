package backend

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"text/template"

	log "github.com/sirupsen/logrus"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/context"

	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/mchudgins/certMgr/pkg/certMgr"
	"github.com/mchudgins/certMgr/pkg/healthz"
	pb "github.com/mchudgins/certMgr/pkg/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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
		log.Debugf("grpcEndpointLog %s+", s)
		defer log.Debugf("grpcEndpointLog %s-", s)
		return handler(ctx, req)
	}
}

// Run the backend command
func Run(cfg *certMgr.AppConfig) {
	server := &server{cfg: *cfg}

	// set the log level
	if cfg.Verbose {
		log.SetLevel(log.DebugLevel)
	}

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}

	// create the Certificate Authority
	server.ca, err = NewCertificateAuthorityFromConfig(cfg)

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

		var s *grpc.Server

		if cfg.Insecure {
			s = grpc.NewServer(
				grpc_middleware.WithUnaryServerChain(
					grpc_prometheus.UnaryServerInterceptor,
					grpcEndpointLog("certMgr")))
		} else {
			tlsCreds, err := credentials.NewServerTLSFromFile(cfg.CertFilename, cfg.KeyFilename)
			if err != nil {
				log.WithError(err).Fatal("Failed to generate grpc TLS credentials")
			}
			s = grpc.NewServer(
				grpc.Creds(tlsCreds),
				grpc.RPCCompressor(grpc.NewGZIPCompressor()),
				grpc.RPCDecompressor(grpc.NewGZIPDecompressor()),
				grpc_middleware.WithUnaryServerChain(
					grpc_prometheus.UnaryServerInterceptor,
					grpcEndpointLog("certMgr")))
		}

		pb.RegisterCertMgrServer(s, server)

		if cfg.Insecure {
			log.Warnf("gRPC service listening insecurely on %s", cfg.GRPCListenAddress)
		} else {
			log.Infof("gRPC service listening on %s", cfg.GRPCListenAddress)
		}
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

		tlsServer := &http.Server{
			Addr: cfg.HTTPListenAddress,
			TLSConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
		}

		// FIXME:  cluster can't health check the self-signed cert endpoint
		if true {
			log.Warnf("HTTP service listening insecurely on %s", cfg.HTTPListenAddress)
			errc <- http.ListenAndServe(cfg.HTTPListenAddress, nil)
		} else {
			log.Infof("HTTPS service listening on %s", cfg.HTTPListenAddress)
			errc <- tlsServer.ListenAndServeTLS(cfg.CertFilename, cfg.KeyFilename)
		}
	}()

	// wait for somthin'
	log.Infof("exit: %s", <-errc)
}
