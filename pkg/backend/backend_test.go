package backend

import (
	"context"
	"log"
	"testing"

	pb "github.com/mchudgins/certMgr/pkg/service"
	"google.golang.org/grpc"
)

var (
	grpcAddr    = ":50051" //flag.String("grpc", ":50051", "listen address for the gRPC server")
	defaultName = "www.example.com"
)

func TestCreate(t *testing.T) {
	// Set up a connection to the server.
	conn, err := grpc.Dial(grpcAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewCertMgrClient(conn)

	// Contact the server and print out its response.
	name := defaultName
	r, err := c.CreateCertificate(context.Background(), &pb.CreateRequest{Name: name, Duration: 90})
	if err != nil {
		log.Fatalf("could not create certificate: %v", err)
	}
	log.Printf("Certificate: %s\nKey: %s", r.GetCertificate(), r.GetKey())
}
