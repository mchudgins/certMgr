package utils

import (
	"os"

	log "github.com/sirupsen/logrus"
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
