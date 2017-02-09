package utils

import (
	"errors"
	"net/http"

	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/afex/hystrix-go/hystrix"
)

type hystrixHelper struct {
	commandName string
}

type writer struct {
	w             http.ResponseWriter
	statusCode    int
	contentLength int
}

/*
func NewLoggingWriter(w http.ResponseWriter) *loggingWriter {
	return &loggingWriter{w: w}
}
*/

func (l *writer) Header() http.Header {
	return l.w.Header()
}

func (l *writer) Write(data []byte) (int, error) {
	l.contentLength += len(data)
	return l.w.Write(data)
}

func (l *writer) WriteHeader(status int) {
	//log.Printf("StatusCode: %d", status)
	l.statusCode = status
	l.w.WriteHeader(status)
}

func (l *writer) Length() int {
	return l.contentLength
}

func (l *writer) StatusCode() int {

	// if nobody set the status, but data has been written
	// then all must be well.
	if l.statusCode == 0 && l.contentLength > 0 {
		return http.StatusOK
	}

	return l.statusCode
}

func NewHystrixHelper(commandName string) (*hystrixHelper, error) {
	hystrix.ConfigureCommand(commandName, hystrix.CommandConfig{
		MaxConcurrentRequests: 100,
	})
	return &hystrixHelper{commandName: commandName}, nil
}

func (y *hystrixHelper) Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := hystrix.Do(y.commandName, func() (err error) {
			//log.Printf("breaker closed")

			monitor := &writer{w: w}
			//log.Printf("monitor.StatusCode: %d", monitor.StatusCode())
			h.ServeHTTP(monitor, r)

			rc := monitor.StatusCode()
			if rc >= 500 && rc < 600 {
				log.WithField("hystrixCommand", y.commandName).
					WithField("StatusCode", rc).Warn("StatusCode indicates backend failure")
				return errors.New(fmt.Sprintf("StatusCode (%d) indicates backend failure", rc))
			}
			return nil
		}, func(err error) error {
			//log.Printf("breaker open")
			log.WithError(err).WithField("hystrixCommand", y.commandName).Warn("hystrix error handler invoked")
			//			w.WriteHeader(http.StatusServiceUnavailable)
			return nil
		})
		if err != nil {
			log.WithError(err).WithField("hystrixCommand", y.commandName).Error("hystrix.Do with error")
		}
	})
}
