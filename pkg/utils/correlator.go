package utils

import (
	"net/http"

	"github.com/rs/xid"
)

// Correlator returns the value of X-Correlation-ID from the HTTP request
type Correlator interface {
	CorrelationID() string
}

// CoreRequest contains the two fields every request should have:
// a correlation ID and a user ID.
type CoreRequest struct {
	txID   string
	userID string
}

// CorrelationID supports the Correlator interface
func (c *CoreRequest) CorrelationID() string {
	return xid.New().String()
}

func NewCoreRequest() *CoreRequest {
	return &CoreRequest{}
}

func (c CoreRequest) CorrelateRequest(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Correlation-ID", c.CorrelationID())
		h.ServeHTTP(w, r)
	})
}
