package backend

import (
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/mchudgins/certMgr/pkg/certMgr"
	"time"
)

func findAndReadFile(fileName string, fileDesc string) (string, error) {
	const fileStr string = "file"

	cfg, err := os.Open(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			log.WithError(err).WithField(fileStr, fileName).Error(fileDesc + " file does not exist.")
		}
		if os.IsPermission(err) {
			log.WithError(err).WithField(fileStr, fileName).Error("Insufficient privilege to read " + fileDesc + ".")
		}
		return "", err
	}
	defer cfg.Close()

	info, err := os.Stat(fileName)
	if err != nil {
		log.WithError(err).WithField(fileStr, fileName).Error("Unable to stat " + fileDesc + " file.")
	}
	buf := make([]byte, info.Size())

	cb, err := cfg.Read(buf)
	if err != nil || int64(cb) != info.Size() {
		log.WithError(err).WithFields(log.Fields{fileStr: fileName, "bytes read": cb, "bytes expected": info.Size()}).
			Error("Unable to read the entire " + fileDesc + " file")
		return "", err
	}

	return string(buf), nil
}

func NewCertificateAuthority(caName string,
	certFile string,
	keyFile string,
	bundleFile string) (*ca, error) {
	cert, err := findAndReadFile(certFile, "certificate")
	if err != nil {
		return nil, err
	}

	key, err := findAndReadFile(keyFile, "key")
	if err != nil {
		return nil, err
	}

	bundle, err := findAndReadFile(bundleFile, "ca bundle")
	if err != nil {
		return nil, err
	}

	return createCA(caName, []byte(cert), []byte(key), bundle)
}

func NewCertificateAuthorityFromConfig(cfg *certMgr.AppConfig) (*ca, error) {
	duration := time.Duration(cfg.Backend.MaxDuration)
	_ = duration

	// find the public portion of the Signing CA
	if len(cfg.Backend.SigningCertificate) == 0 {

	}
	return createCA("", []byte(""), []byte(""), "")
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
