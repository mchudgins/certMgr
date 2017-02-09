package backend

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"net"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"golang.org/x/net/context"

	"os"

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

// CreateCertificate creates an x509 certificate
func (s *server) CreateCertificate(ctx context.Context, in *pb.CreateRequest) (*pb.CreateReply, error) {

	md, _ := metadata.FromContext(ctx)
	for key, value := range md {
		log.Debugf("md[ %s ] : %s", key, value[0])
	}

	var validFor time.Duration
	validFor = time.Duration(in.GetDuration()) * time.Hour * 24

	cert, key, err := s.ca.CreateCertificate(ctx, in.GetName(), in.GetAlternateNames(), validFor)
	return &pb.CreateReply{Certificate: cert, Key: key}, err
}

func (c ca) CreateCertificate(ctx context.Context,
	commonName string,
	alternateNames []string,
	duration time.Duration) (cert string, key string, err error) {
	requestedHosts := make([]string, len(alternateNames)+1, len(alternateNames)+1)
	requestedHosts[0] = commonName
	copy(requestedHosts[1:], alternateNames)

	hosts, err := c.validateRequest(requestedHosts, duration)
	if err != nil {
		return "", "", err
	}

	// create the CSR

	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	/*
		keyOut, err := os.Create("key.pem")
		defer keyOut.Close()
		if err != nil {
			log.Fatalf("failed to open key.pem for writing: %s", err)
		}
		pem.Encode(keyOut, pemBlockForKey(priv))
	*/

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	notBefore := time.Now()
	notAfter := notBefore.Add(duration)
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   hosts[0],
			Organization: []string{"DST Systems, Inc"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage: x509.KeyUsageKeyEncipherment |
			x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	// sign the CSR
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &c.SigningCertificate, publicKey(priv), c.SigningKey)
	if err != nil {
		log.WithError(err).Error("Unable to CreateCertificate")
		return "", "", err
	}

	// prepare the response
	var certBuffer, keyBuffer bytes.Buffer

	/*
		certOut, err := os.Create("cert.pem")
		defer certOut.Close()
		if err != nil {
			log.WithError(err).Fatal("failed to open cert.pem for writing")
		}

		pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	*/

	pem.Encode(&certBuffer, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	cert = certBuffer.String()
	pem.Encode(&keyBuffer, pemBlockForKey(priv))
	key = keyBuffer.String()

	//	// persist the certificate
	//	serverCert, err := x509.ParseCertificate(derBytes)
	//	persistedCert := newCertFromCertificate(serverCert)
	//
	//	/* using hystrix/circuitbreaker to persist the data */
	//	err = hystrix.Do("certs-mysql", func() error {
	//		persistedCert.Insert()
	//		return nil
	//	}, nil)

	/*
		data := &CertificateData{*serverCert}
		data.Persist(ctx)
	*/

	return cert, key, err
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

func publicKey(priv interface{}) interface{} {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	default:
		return nil
	}
}

func pemBlockForKey(priv interface{}) *pem.Block {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(k)}
	case *ecdsa.PrivateKey:
		b, err := x509.MarshalECPrivateKey(k)
		if err != nil {
			log.WithError(err).Fatal("Unable to marshall ECDSA private key")
			os.Exit(2)
		}
		return &pem.Block{Type: "EC PRIVATE KEY", Bytes: b}
	default:
		return nil
	}
}

func (c ca) String() string {
	return "Name: " + c.Name + "; ...."
}
