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
	"time"

	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/metrics"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/mchudgins/golang-backend-starter/healthz"
	"github.com/mchudgins/golang-backend-starter/utils"
	"google.golang.org/grpc"
)

type server struct{}

var (
	fieldKeys    = []string{"method", "error"}
	requestCount = kitprometheus.NewCounter(stdprometheus.CounterOpts{
		Namespace: "my_group",
		Subsystem: "cert_service",
		Name:      "request_count",
		Help:      "Number of requests received.",
	}, fieldKeys)
	requestLatency = metrics.NewTimeHistogram(time.Microsecond, kitprometheus.NewSummary(stdprometheus.SummaryOpts{
		Namespace: "my_group",
		Subsystem: "cert_service",
		Name:      "request_latency_microseconds",
		Help:      "Total duration of requests in microseconds.",
	}, fieldKeys))
	countResult = kitprometheus.NewSummary(stdprometheus.SummaryOpts{
		Namespace: "my_group",
		Subsystem: "cert_service",
		Name:      "count_result",
		Help:      "The result of each count method.",
	}, []string{}) // no fields here
	counters = instrumentation{requestCount, requestLatency, countResult}

	helloEndpoint = endpoint.Chain(endpointInstrumentation(&counters, "SayHello"),
		endpointLog("SayHello"))(hello)

	// boilerplate variables for good SDLC hygiene.  These are auto-magically
	// injected by the Makefile & linker working together.
	version   string
	buildTime string
	builder   string
	goversion string
)

func hello(ctx context.Context, req interface{}) (resp interface{}, err error) {
	log.Printf("in hello3()")
	return &HelloReply{Message: "Hello " + "fubar"}, nil
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *HelloRequest) (*HelloReply, error) {
	resp, err := helloEndpoint(ctx, in)
	return resp.(*HelloReply), err
}

type instrumentation struct {
	requestCount   metrics.Counter
	requestLatency metrics.TimeHistogram
	countResult    metrics.Histogram
}

func endpointInstrumentation(i *instrumentation, s string) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			log.Printf("endpointInstrumentation+")
			defer log.Printf("endpointInstrumentation-")

			defer func(begin time.Time) {
				methodField := metrics.Field{Key: "method", Value: s}
				errorField := metrics.Field{Key: "error", Value: fmt.Sprintf("%v", err)}
				i.requestCount.With(methodField).With(errorField).Add(1)
				i.requestLatency.With(methodField).With(errorField).Observe(time.Since(begin))
			}(time.Now())

			response, err = next(ctx, request)

			return
		}
	}
}

func endpointLog(s string) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			log.Printf("endpointLog %s+", s)
			defer log.Printf("endpointLog %s-", s)

			//			BaseLogger.Log("level", "debug", "msg", fmt.Sprintf("Type of request: %T", request))

			var txid string

			if c, fOK := request.(Correlator); fOK {
				txid = c.CorrelationID()
			} else if req, fOK := request.(*http.Request); fOK {
				txid = req.Header.Get("X-Correlation-ID")
			} else {
				log.Printf("no CorrelationID")
			}

			if len(txid) == 0 {
				log.Printf("Request is missing 'X-Correlation-ID' header: %+v", request)
			}

			ctx = context.WithValue(ctx, "txid", txid)

			return next(ctx, request)
		}
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

		s := grpc.NewServer()
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
