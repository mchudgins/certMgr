package utils

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/afex/hystrix-go/hystrix"
)

// FindAndReadFile loads the specified file from disk
func FindAndReadFile(fileName string, fileDesc string) (string, error) {
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

// readFile
func readFile(uri string) (io.ReadCloser, error) {

	// did they include a "file://"?
	filename := uri
	if strings.HasPrefix(uri, "file://") {
		filename = uri[0:len("file://")]
	}

	file, err := os.OpenFile(filename, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}

	return file, nil
}

// readViaNet
func readViaNet(uri string) (io.ReadCloser, error) {

	c := &http.Client{}

	url, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	var resp *http.Response

	err = hystrix.Do(url.Host, func() (err error) {
		r, err := c.Get(uri)
		if err != nil {
			return err
		}
		resp = r
		return nil
	}, nil)

	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

// readConfig
func OpenReadCloser(uri string) (io.ReadCloser, error) {

	switch uri[0:5] {
	case "http:":
		return readViaNet(uri)

	case "https":
		return readViaNet(uri)

	case "file:":
		return readFile(uri[7:])

	default:
		log.Printf("unable to interpret %s as a file or network location.", uri)
	}

	return nil, errors.New("unable to determine access mode")
}
