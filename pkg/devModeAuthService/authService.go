package devModeAuthService

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	log "github.com/Sirupsen/logrus"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/mchudgins/certMgr/pkg/healthz"
	pb "github.com/mchudgins/certMgr/pkg/service"
	"github.com/mchudgins/certMgr/pkg/utils"
	"github.com/mwitkow/go-grpc-middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cobra"
)

type server struct{}

func (s *server) Configuration(ctx context.Context,
	in *pb.ConfigurationRequest) (*pb.ConfigurationResponse, error) {

	resp := &pb.ConfigurationResponse{
		LogonURL:  "http://localhost:9999/signin",
		LogoutURL: "http://localhost:9999/logout",
	}

	return resp, nil
}

func (s *server) VerifyToken(ctx context.Context,
	in *pb.VerificationRequest) (*pb.VerificationResponse, error) {
	resp := &pb.VerificationResponse{
		Valid:           true,
		UserID:          in.Token,
		CacheExpiration: time.Now().Add(15 * time.Minute).Unix(),
	}

	return resp, nil
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

func Command(cmd *cobra.Command, args []string) {
	// TODO: Work your own magic here
	log.Info("'authService' started!  This command is for Development mode ONLY!")

	cfg, err := utils.NewAppConfig(cmd)
	if err != nil {
		log.WithError(err).Fatal("Unable to initialize the application.  Exiting now.")
	}

	// if we were given a cert, write it out to a tmp file
	if len(cfg.Certificate) != 0 {
		tmpfile, err := ioutil.TempFile("", "be")
		if err != nil {
			log.WithError(err).
				Fatal("unable to create a tmp file for certificate")
		}

		cfg.CertFilename = tmpfile.Name()

		if _, err = tmpfile.Write([]byte(cfg.Certificate)); err != nil {
			log.WithError(err).
				Fatalf("an error occurred while writing the certificate to temporary file %s",
					tmpfile.Name())
		}
		if err = tmpfile.Close(); err != nil {
			log.WithError(err).
				Fatalf("an error occurred while closing the temporary file %s",
					tmpfile.Name())
		}
	}

	utils.StartUpMessage(*cfg)

	listenAddress := cfg.AuthServiceAddress

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
		lis, err := net.Listen("tcp", listenAddress)
		if err != nil {
			errc <- err
			return
		}

		var s *grpc.Server

		if cfg.Insecure {
			s = grpc.NewServer(
				grpc_middleware.WithUnaryServerChain(
					grpc_prometheus.UnaryServerInterceptor,
					grpcEndpointLog("devModeAuthServer")))
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
					grpcEndpointLog("devModeAuthServer")))
		}

		pb.RegisterAuthVerifierServer(s, &server{})

		if cfg.Insecure {
			log.Warnf("gRPC service listening insecurely on %s", listenAddress)
		} else {
			log.Infof("gRPC service listening on %s", listenAddress)
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
		})

		log.Infof("HTTP service listening on %s", cfg.HTTPListenAddress)
		errc <- http.ListenAndServe(cfg.HTTPListenAddress, nil)
	}()

	// wait for somthin'
	log.Infof("exit: %s", <-errc)
}
