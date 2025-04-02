package backend

import (
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"strings"

	"time"

	log "github.com/sirupsen/logrus"
	"github.com/mchudgins/certMgr/pkg/assets"
	"github.com/mchudgins/certMgr/pkg/certMgr"
	"github.com/mchudgins/certMgr/pkg/utils"
)

func NewCertificateAuthority(caName string,
	certFile string,
	keyFile string,
	bundleFile string) (*ca, error) {
	cert, err := utils.FindAndReadFile(certFile, "certificate")
	if err != nil {
		return nil, err
	}

	key, err := utils.FindAndReadFile(keyFile, "key")
	if err != nil {
		return nil, err
	}

	bundle, err := utils.FindAndReadFile(bundleFile, "ca bundle")
	if err != nil {
		return nil, err
	}

	return createCA(caName, []byte(cert), []byte(key), bundle)
}

func loadAsset(asset string) (string, error) {
	b, err := assets.Asset(asset)
	if err != nil {
		log.WithError(err).WithField("asset", asset)
		return "", err
	}
	return string(b), nil
}

func NewCertificateAuthorityFromConfig(cfg *certMgr.AppConfig) (*ca, error) {
	var err error
	duration := time.Duration(cfg.Backend.MaxDuration)
	_ = duration

	// find the public portion of the Signing CA
	cert := cfg.Backend.SigningCACertificate
	if len(cert) == 0 {
		cert, err = loadAsset("static/signing-ca.crt")
		if err != nil {
			log.WithError(err).Fatal("Application misconfigured, exiting.")
		}
	}

	// find the bundle of intermediate CA's
	bundle := cfg.Backend.Bundle
	if len(bundle) == 0 {
		bundle, err = loadAsset("static/ca-bundle.pem")
		if err != nil {
			log.WithError(err).Fatal("Application misconfigured, exiting.")
		}
	}

	key, err := utils.FindAndReadFile(cfg.Backend.SigningCAKeyFilename, "CA key")
	if err != nil {
		log.WithError(err).Fatalf("Application misconfigured, exiting")
	}

	return createCA("", []byte(cert), []byte(key), bundle)
}

func createCA(caName string,
	cert []byte,
	key []byte,
	bundle string) (*ca, error) {

	if len(caName) == 0 {
		caName = "default"
	}

	pemCert, _ := pem.Decode(cert)
	if pemCert == nil {
		msg := "Unable to decode the certificate!"
		log.Error(msg)
		return nil, errors.New(msg)
	}

	pemKey, _ := pem.Decode(key)
	if pemKey == nil {
		msg := "Unable to decode the certificate's key!"
		log.Error(msg)
		return nil, errors.New(msg)
	}

	if x509.IsEncryptedPEMBlock(pemKey) {
		msg := "Certificate key requires a passphrase! This is unsupported."
		log.Error(msg)
		return nil, errors.New(msg)
	}
	caKey, err := x509.ParsePKCS8PrivateKey(pemKey.Bytes)
	if err != nil {
		msg := "Unable to parse certificate's key"
		log.WithError(err).Error(msg)
		return nil, errors.New(msg)
	}

	if _, ok := caKey.(crypto.Signer); !ok {
		msg := "hmmm, the CA private key is not a crypto.Signer"
		log.Error(msg)
		return nil, errors.New(msg)
	}

	caCertificate, err := x509.ParseCertificate(pemCert.Bytes)
	if err != nil {
		log.WithError(err).Error("error parsing CA certificate")
	}

	log.Infof("permittedDomains:  %s", strings.Join(caCertificate.PermittedDNSDomains, ", "))

	return &ca{Name: caName,
		SigningCertificate: *caCertificate,
		SigningKey:         caKey.(crypto.Signer),
		Bundle:             bundle}, nil
}
