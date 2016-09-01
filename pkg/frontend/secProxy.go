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

		h.ServeHTTP(w, r)
	})
}
