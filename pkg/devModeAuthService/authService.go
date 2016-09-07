package devModeAuthService

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/grpc"

	"github.com/grpc-ecosystem/go-grpc-prometheus"
	pb "github.com/mchudgins/golang-service-starter/pkg/service"
	"github.com/mchudgins/golang-service-starter/pkg/utils"
	"github.com/mwitkow/go-grpc-middleware"
	"github.com/spf13/cobra"
)

type server struct{}

func (s *server) Configuration(ctx context.Context,
	in *pb.ConfigurationRequest) (*pb.ConfigurationResponse, error) {

	resp := &pb.ConfigurationResponse{
		LogonURL:  "http://localhost:9999/sigin",
		LogoutURL: "http://localhost:9999/logout",
	}

	return resp, nil
}

func (s *server) VerifyToken(ctx context.Context,
	in *pb.VerificationRequest) (*pb.VerificationResponse, error) {
	resp := &pb.VerificationResponse{
		Valid:           true,
		UserID:          in.Token,
		CacheExpiration: int64(15 * time.Minute),
	}

	return resp, nil
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

func Command(cmd *cobra.Command, args []string) {
	// TODO: Work your own magic here
	log.Printf("'authService' started!  This command is for Development mode ONLY!")

	cfg, err := utils.NewAppConfig(cmd)
	if err != nil {
		log.Printf("Unable to initialize the application (%s).  Exiting now.", err)
	}

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

		s := grpc.NewServer(
			grpc_middleware.WithUnaryServerChain(
				grpc_prometheus.UnaryServerInterceptor,
				grpcEndpointLog("devModeAuthServer")))
		pb.RegisterAuthVerifierServer(s, &server{})
		log.Printf("gRPC service listening on %s", listenAddress)
		errc <- s.Serve(lis)
	}()

	// wait for somthin'
	log.Printf("exit: %s", <-errc)
}
