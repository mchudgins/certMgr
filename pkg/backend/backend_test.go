package backend

import (
	"context"
	"testing"

	log "github.com/Sirupsen/logrus"
	pb "github.com/mchudgins/certMgr/pkg/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	grpcAddr    = ":50051" //flag.String("grpc", ":50051", "listen address for the gRPC server")
	defaultName = "fubar.cap.dstcorp.io"
)

func TestCreate(t *testing.T) {
	// Set up a connection to the server.
	conn, err := grpc.Dial(defaultName+grpcAddr, grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(nil, "")))
	if err != nil {
		log.WithError(err).Fatal("did not connect")
	}
	defer conn.Close()
	c := pb.NewCertMgrClient(conn)

	// Contact the server and print out its response.
	name := defaultName
	r, err := c.CreateCertificate(context.Background(), &pb.CreateRequest{Name: name, Duration: 90})
	if err != nil {
		log.WithError(err).Fatalf("could not create certificate: %v", err)
	}
	log.Printf("Certificate: %s\nKey: %s", r.GetCertificate(), r.GetKey())
}
