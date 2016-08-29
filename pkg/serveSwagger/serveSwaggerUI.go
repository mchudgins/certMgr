package serveSwagger

import (
	"log"
	"net/http"
)

// This package serves up the Swagger UI at the designated path
// example:
//  		swaggerProxy, _ := serveSwagger.NewSwaggerProxy("/swagger-ui/")
//  		http.HandleFunc("/swagger-ui/", swaggerProxy.ServeHTTP)
//      http.ListenAndServe(":8080", nil)

// SwaggerProxy serves the swagger UI at the designated path
type SwaggerProxy struct {
	path    string
	pathLen int
	h       http.Handler
}

// NewSwaggerProxy initializes the SwaggerProxy struct
func NewSwaggerProxy(proxyPath string) (*SwaggerProxy, error) {
	return &SwaggerProxy{path: proxyPath,
		pathLen: len(proxyPath),
		h:       http.FileServer(http.Dir("/home/mchudgins/golang/src/github.com/swagger-api/swagger-ui/dist"))}, nil
}

// ServeHTTP serves up the Swagger UI
func (s *SwaggerProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	log.Printf("Query: %s", r.URL.RawQuery)
	path := r.URL.Path[s.pathLen:]
	if len(path) == 0 && len(r.URL.RawQuery) == 0 {
		http.Redirect(w, r,
			"/swagger-ui/?url=/swagger/service.swagger.json", http.StatusMovedPermanently)
		return
	}

	r.URL.Path = path
	s.h.ServeHTTP(w, r)
}
