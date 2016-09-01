package frontend

import (
	"net/http"
	"net/url"
)

type securityProxy struct {
	url url.URL
}

func NewSecurityProxy(idp string) (securityProxy, error) {

	url, err := url.Parse(idp)
	if err != nil {
		return securityProxy{}, err
	}

	return securityProxy{url: *url}, nil
}

func (s *securityProxy) Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if auth := r.Header.Get("Authorization"); len(auth) == 0 {
			http.Redirect(w, r,
				s.url.String(), http.StatusTemporaryRedirect)
			return
		}

		// Todo:  if the authorization header is present, then determine the
		// user's ID and pass it along to the backend via the grpc per call credentials
		h.ServeHTTP(w, r)
	})
}
