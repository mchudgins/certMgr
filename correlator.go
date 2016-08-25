package main

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
func (r CoreRequest) CorrelationID() string {
	return r.txID
}
