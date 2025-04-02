package backend

import (
	"context"
	"testing"

	"os"

	log "github.com/sirupsen/logrus"
	pb "github.com/mchudgins/certMgr/pkg/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	grpcAddr    = ":50051" //flag.String("grpc", ":50051", "listen address for the gRPC server")
	defaultName = "fubar.cap.dstcorp.io"
)

func createConnection() (pb.CertMgrClient, *grpc.ClientConn, error) {
	// Set up a connection to the server.
	conn, err := grpc.Dial(defaultName+grpcAddr, //grpc.WithInsecure())
		grpc.WithCompressor(grpc.NewGZIPCompressor()),
		grpc.WithDecompressor(grpc.NewGZIPDecompressor()),
		grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(nil, "")))

	if err != nil {
		log.WithError(err).Fatal("did not connect")
	}

	//	defer conn.Close()
	c := pb.NewCertMgrClient(conn)

	return c, conn, nil
}

func TestCreate(t *testing.T) {
	c, conn, err := createConnection()
	if err != nil {
		log.WithError(err).Fatal("unable to contact server")
	}
	defer conn.Close()

	// Contact the server and print out its response.
	name := defaultName
	r, err := c.CreateCertificate(context.Background(), &pb.CreateRequest{Name: name, Duration: 90})
	if err != nil {
		log.WithError(err).Fatalf("could not create certificate: %v", err)
	}
	log.Printf("Certificate: %s\nKey: %s", r.GetCertificate(), r.GetKey())
}

func TestLotsOfCreates(t *testing.T) {
	c, conn, err := createConnection()
	if err != nil {
		log.WithError(err).Fatal("unable to contact server")
	}
	defer conn.Close()

	const loopers int = 1000

	countc := make(chan int, loopers)

	// Contact the server and print out its response.
	name := defaultName

	for i := 0; i < loopers; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				_, err := c.CreateCertificate(context.Background(), &pb.CreateRequest{Name: name, Duration: 90})
				if err != nil {
					log.WithError(err).Fatalf("could not create certificate: %v", err)
				}
			}
			countc <- 0
		}()
	}

	j := 0
	for range countc {
		j++
		if j == loopers {
			os.Exit(0)
		}
	}

}
