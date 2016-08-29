package serveSwagger

import (
	"log"
	"net/http"
)

// ServeHTTP serves up the Swagger UI
func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf(r.URL.Path)
}
