package backend

import (
	"crypto"
	"crypto/x509"
	"errors"
	"log"
	"net"
	"strings"
	"time"

	"golang.org/x/net/context"

	pb "github.com/mchudgins/certMgr/pkg/service"
	"google.golang.org/grpc/metadata"
)

type ca struct {
	Name               string
	SigningCertificate x509.Certificate
	SigningKey         crypto.Signer
	RootCertificate    x509.Certificate
	Bundle             string
}

var SimpleCA *ca

// CreateCertificate creates an x509 certificate
func (s *server) CreateCertificate(ctx context.Context, in *pb.CreateRequest) (*pb.CreateReply, error) {
	log.Printf("ctx: %+v", ctx)

	md, _ := metadata.FromContext(ctx)
	for key, value := range md {
		log.Printf("md[ %s ] : %s", key, value[0])
	}

	log.Printf("common name:  %s", in.GetName())
	var validFor time.Duration
	validFor = time.Duration(in.GetDuration()) * time.Hour * 24
	log.Printf("duration:  %f days", validFor.Hours()/24)
	for _, s := range in.GetAlternateNames() {
		log.Printf("alt:  %s", s)
	}

	return &pb.CreateReply{}, nil
}

func (c *ca) validateRequest(requestedHosts []string, validFor time.Duration) ([]string, error) {
	var hosts = make([]string, len(requestedHosts))

	for i, s := range requestedHosts {
		h := strings.ToLower(s)
		hosts[i] = h

		if strings.HasPrefix(h, "www.") {
			return nil, errors.New("www. host names are not supported")
		}

		if strings.HasPrefix(h, ".") {
			return nil, errors.New(". host names are not supported")
		}

		supportedDomain := false
		for _, dns := range c.SigningCertificate.PermittedDNSDomains {
			tld := "." + dns
			if strings.HasSuffix(h, tld) {
				supportedDomain = true
				break
			} else if ip := net.ParseIP(h); ip != nil && i != 0 {
				supportedDomain = true
			}
		}
		if !supportedDomain {
			return nil, errors.New("one or more host names are not within the permitted domain list")
		}
	}

	return hosts, nil
}

func (c ca) String() string {
	return "Name: " + c.Name + "; ...."
}
