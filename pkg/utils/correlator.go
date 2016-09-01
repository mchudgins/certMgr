package utils

import (
	"net/http"

	"github.com/rs/xid"
)

const (
	XCorrID = "X-Correlation-ID"
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
		id := r.Header.Get(XCorrID)
		if len(id) == 0 {
			id = c.CorrelationID()
		}
		w.Header().Set(XCorrID, id)
		h.ServeHTTP(w, r)
	})
}
