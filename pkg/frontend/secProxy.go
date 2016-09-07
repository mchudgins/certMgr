package frontend

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"

	"google.golang.org/grpc"

	"github.com/afex/hystrix-go/hystrix"
	pb "github.com/mchudgins/golang-service-starter/pkg/service"
)

const (
	remoteUserHeader         = "X-RemoteUser"
	grpcMetadataHeaderPrefix = "Grpc-Metadata-"
)

var (
	bearerRegex  = regexp.MustCompile(`^\s*bearer\s+([[:alnum:]]+)`)
	authVerifier pb.AuthVerifierClient
)

type securityProxy struct {
	url       string
	auth      pb.AuthVerifierClient
	logonURL  string
	logoutURL string
}

func init() {
}

func NewSecurityProxy(idp string) (*securityProxy, error) {

	conn, err := grpc.Dial(idp, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	client := pb.NewAuthVerifierClient(conn)
	server := &securityProxy{url: idp, auth: client}

	err = hystrix.Do(server.url, func() (err error) {

		resp, err := client.Configuration(context.Background(),
			&pb.ConfigurationRequest{})
		if err != nil {
			newerr := fmt.Errorf("Unable to retrieve configuration URLs from the authentication service (%s) -- %s", idp, err)
			return newerr
		}
		server.logonURL = resp.LogonURL
		server.logoutURL = resp.LogoutURL

		return nil
	}, nil)

	log.Printf("server: %+v", server)
	return server, err
}

// DRY: make sure we always redirect to LogonURL in the same way
func (s *securityProxy) redirectToLogon(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r,
		s.logonURL, http.StatusTemporaryRedirect)
}

func (s *securityProxy) Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var token string

		auth := r.Header.Get("Authorization")
		if len(auth) == 0 {
			s.redirectToLogon(w, r)
			return
		}

		if bearerRegex.MatchString(auth) {
			result := bearerRegex.FindStringSubmatch(auth)
			if len(result) == 2 {
				token = result[1]
			}
		}

		// Todo:  if the authorization header is present, then determine the
		// user's ID and pass it along to the backend via the grpc per call credentials
		if len(token) == 0 {
			s.redirectToLogon(w, r)
			return
		}

		var resp *pb.VerificationResponse

		err := hystrix.Do(s.url, func() (err error) {
			request := &pb.VerificationRequest{Token: token}
			resp, err = s.auth.VerifyToken(context.Background(), request)
			if err != nil {
				log.Printf("VerifyToken:  %s", err)
			}
			log.Printf("Response: %+v", resp)

			return nil
		}, nil)

		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		// resp.UserID needs to be passed along to the backend
		r.Header.Set(grpcMetadataHeaderPrefix+remoteUserHeader, resp.UserID)

		// finally, pass the request along the processing chain
		h.ServeHTTP(w, r)
	})
}
