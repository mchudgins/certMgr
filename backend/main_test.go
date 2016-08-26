package main

import (
	"context"
	"log"
	"testing"

	pb "github.com/mchudgins/golang-service-starter/service"
	"google.golang.org/grpc"
)

var (
	grpcAddr    = ":50051" //flag.String("grpc", ":50051", "listen address for the gRPC server")
	defaultName = "world"
)

func TestHello(t *testing.T) {
	// Set up a connection to the server.
	conn, err := grpc.Dial(grpcAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	// Contact the server and print out its response.
	name := defaultName
	r, err := c.SayHello(context.Background(), &pb.HelloRequest{Name: name})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", r.Message)
}
