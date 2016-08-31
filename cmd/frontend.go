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
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/mchudgins/go-service-helper/pkg/loggingWriter"
	"github.com/mchudgins/go-service-helper/pkg/serveSwagger"
	"github.com/mchudgins/golang-service-starter/pkg/healthz"
	pb "github.com/mchudgins/golang-service-starter/pkg/service"
	"github.com/mchudgins/golang-service-starter/pkg/utils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

// frontendCmd represents the frontend command
var (
	swagger     = MustAsset("pkg/service/service.swagger.json")
	frontendCmd = &cobra.Command{
		Use:   "frontend",
		Short: "launch a 'frontend' for the sample service",
		Long: `Launch a 'frontend' API server for the sample service.

The frontend service accepts and provides JSON for the gRPC backend.`,
		Run: frontend,
	}
)

func init() {
	RootCmd.AddCommand(frontendCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// frontendCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// frontendCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}

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

func frontend(cmd *cobra.Command, args []string) {

	cfg, err := utils.NewAppConfig(cmd)
	if err != nil {
		log.Printf("Unable to initialize the application (%s).  Exiting now.", err)
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
