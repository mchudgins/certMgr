// Copyright © 2016 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

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
	"github.com/mchudgins/golang-service-starter/pkg/healthz"
	pb "github.com/mchudgins/golang-service-starter/pkg/service"
	"github.com/mchudgins/golang-service-starter/pkg/utils"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

// backendCmd represents the backend command
var backendCmd = &cobra.Command{
	Use:   "backend",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: backend,
}

func init() {
	RootCmd.AddCommand(backendCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// backendCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// backendCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}

type server struct{}

var (
	// boilerplate variables for good SDLC hygiene.  These are auto-magically
	// injected by the Makefile & linker working together.
	version   string
	buildTime string
	builder   string
	buildNum  string
	goversion string
)

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("ctx: %+v", ctx)
	return &pb.HelloReply{Message: "Hello " + in.Name}, nil
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

func backend(cmd *cobra.Command, args []string) {
	log.Printf("golang-backend-starter: version %s; buildTime: %s; built by: %s; buildNum: %s; (%s)",
		version, buildTime, builder, buildNum, goversion)

	cfg, err := utils.NewAppConfig(cmd)
	if err != nil {
		log.Printf("Unable to initialize the application (%s).  Exiting now.", err)
	}

	log.Printf("Starting app ...")

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}

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
		pb.RegisterGreeterServer(s, &server{})
		log.Printf("gRPC service listening on %s", cfg.GRPCListenAddress)
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
				fmt.Fprintf(w, "Unable to parse template: %s", err)
				return
			}

			err = tmp.Execute(w, data{Hostname: hostname})
			if err != nil {
				fmt.Fprintf(w, "Unable to execute template: %s", err)
			}
		})

		log.Printf("HTTP service listening on %s", cfg.HTTPListenAddress)
		errc <- http.ListenAndServe(cfg.HTTPListenAddress, nil)
	}()

	// wait for somthin'
	log.Printf("exit: %s", <-errc)
}
