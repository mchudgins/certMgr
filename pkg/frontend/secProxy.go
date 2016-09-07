package frontend

import (
	"context"
	"log"
	"net/http"
	"regexp"

	"google.golang.org/grpc"

	pb "github.com/mchudgins/golang-service-starter/pkg/service"
)

var (
	bearerRegex  = regexp.MustCompile(`^\s*bearer\s+([[:alnum:]]+)`)
	authVerifier pb.AuthVerifierClient
)

type securityProxy struct {
	url  string
	auth pb.AuthVerifierClient
}

func init() {
}

func NewSecurityProxy(idp string) (*securityProxy, error) {

	conn, err := grpc.Dial(idp, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	client := pb.NewAuthVerifierClient(conn)

	return &securityProxy{url: idp, auth: client}, nil
}

func (s *securityProxy) Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var token string

		auth := r.Header.Get("Authorization")
		if len(auth) == 0 {
			http.Redirect(w, r,
				s.url, http.StatusTemporaryRedirect)
			return
		}

		log.Printf("auth: %s", auth)
		if bearerRegex.MatchString(auth) {
			log.Printf("its a match")
			result := bearerRegex.FindStringSubmatch(auth)
			log.Printf("token: %s", result)
			log.Printf("%d items in result", len(result))
			if len(result) == 2 {
				log.Printf("token: %s", result[1])
				token = result[1]
			}
		}

		// Todo:  if the authorization header is present, then determine the
		// user's ID and pass it along to the backend via the grpc per call credentials
		if len(token) == 0 {
			http.Redirect(w, r,
				s.url, http.StatusTemporaryRedirect)
			return
		}

		request := &pb.VerificationRequest{Token: token}
		resp, err := s.auth.VerifyToken(context.Background(), request)
		if err != nil {
			log.Printf("VerifyToken:  %s", err)
		}
		log.Printf("Response: %+v", resp)

		h.ServeHTTP(w, r)
	})
}
