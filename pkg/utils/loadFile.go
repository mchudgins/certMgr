package utils

import (
	"errors"
	"io"
	"net/http"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
)

// FindAndReadFile loads the specified file from disk
func FindAndReadFile(fileName string, fileDesc string) (string, error) {
	const fileStr string = "file"

	cfg, err := os.OpenFile(fileName, os.O_RDONLY, 0)
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

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	return file, nil
}

// readViaNet
func readViaNet(uri string, desc string) (io.ReadCloser, error) {

	c := &http.Client{}

	return c.Get(uri)
}

// readConfig
func OpenReadCloser(uri string, desc string) (io.ReadCloser, error) {

	switch uri[0:5] {
	case "http:":
		return readConfigViaNet(uri)

	case "https":
		return readConfigViaNet(uri)

	case "file:":
		return readConfigFile(uri[7:])

	default:
		log.Printf("Warning: unable to interpret %s as a file or network location.", uri)
	}

	return nil, errors.New("unable to determine access mode")
}
