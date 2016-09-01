package utils

import (
	"fmt"
	"log"
	"net/http"

	"github.com/afex/hystrix-go/hystrix"
)

type hystrixHelper struct {
	commandName   string
	w             http.ResponseWriter
	statusCode    int
	contentLength int
}

/*
func NewLoggingWriter(w http.ResponseWriter) *loggingWriter {
	return &loggingWriter{w: w}
}
*/

func (l *hystrixHelper) Header() http.Header {
	return l.w.Header()
}

func (l *hystrixHelper) Write(data []byte) (int, error) {
	l.contentLength += len(data)
	return l.w.Write(data)
}

func (l *hystrixHelper) WriteHeader(status int) {
	l.statusCode = status
	l.w.WriteHeader(status)
}

func (l *hystrixHelper) Length() int {
	return l.contentLength
}

func (l *hystrixHelper) StatusCode() int {

	// if nobody set the status, but data has been written
	// then all must be well.
	if l.statusCode == 0 && l.contentLength > 0 {
		return http.StatusOK
	}

	return l.statusCode
}

func NewHystrixHelper(commandName string) (*hystrixHelper, error) {
	return &hystrixHelper{commandName: commandName}, nil
}

func (y *hystrixHelper) Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := hystrix.Do(y.commandName, func() (err error) {
			y.w = w
			h.ServeHTTP(y, r)
			rc := y.StatusCode()
			if rc >= 500 && rc < 600 {
				return fmt.Errorf("backend failure")
			}
			return nil
		}, func(err error) error {
			log.Printf("hystrix error handler for command %s with error %s", y.commandName, err)
			w.WriteHeader(http.StatusServiceUnavailable)
			return nil
		})
		log.Printf("hystrix.Handler with error: %s", err)
	})
}
