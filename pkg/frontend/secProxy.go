package frontend

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"google.golang.org/grpc"

	log "github.com/Sirupsen/logrus"
	"github.com/afex/hystrix-go/hystrix"
	pb "github.com/mchudgins/certMgr/pkg/service"
	"github.com/patrickmn/go-cache"
)

const (
	remoteUserHeader         = "X-RemoteUser"
	grpcMetadataHeaderPrefix = "Grpc-Metadata-"
)

var (
	bearerRegex  = regexp.MustCompile(`^\s*bearer\s+([[:alnum:]]+)`)
	authVerifier pb.AuthVerifierServiceClient
)

type securityProxy struct {
	url       string
	auth      pb.AuthVerifierServiceClient
	logonURL  string
	logoutURL string
	cache     *cache.Cache
}

func init() {
}

func NewSecurityProxy(idp string) (*securityProxy, error) {

	conn, err := grpc.Dial(idp, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	client := pb.NewAuthVerifierServiceClient(conn)
	server := &securityProxy{url: idp, auth: client}

	hystrix.ConfigureCommand(server.url, hystrix.CommandConfig{
		Timeout:               250, // 250 ms
		MaxConcurrentRequests: 100,
	})

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

	// set up the cache
	server.cache = cache.New(30*time.Minute, 30*time.Second)

	return server, err
}

// DRY: make sure we always redirect to LogonURL in the same way
func (s *securityProxy) redirectToLogon(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r,
		s.logonURL, http.StatusTemporaryRedirect)
}

func (s *securityProxy) verifyToken(token string) (*pb.VerificationResponse, error) {
	var resp *pb.VerificationResponse

	err := hystrix.Do(s.url, func() (err error) {

		request := &pb.VerificationRequest{Token: token}
		resp, err = s.auth.VerifyToken(context.Background(), request)
		if err != nil {
			log.WithError(err).WithField("token", token).Warn("VerifyToken returned an error")
		}

		return err
	}, nil)

	return resp, err
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

		// check the process cache to see if the token is valid
		now := time.Now()
		cacheHit, found := s.cache.Get(token)
		if !found { // not in the cache, go get it
			var err error
			resp, err = s.verifyToken(token)

			if err != nil {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}

			expires := time.Unix(resp.CacheExpiration, 0)
			if resp.Valid && expires.After(now) {
				s.cache.Set(token, resp, expires.Sub(now))
			}
		} else { // in the cache, double check it
			resp = cacheHit.(*pb.VerificationResponse)
			expires := time.Unix(resp.CacheExpiration, 0)
			if now.After(expires) {
				resp.Valid = false
				s.cache.Delete(token)
			}
		}

		// if the token's invalid, send 'em to the logon URL
		if !resp.Valid {
			s.redirectToLogon(w, r)
			return
		}

		// resp.UserID needs to be passed along to the backend
		r.Header.Set(grpcMetadataHeaderPrefix+remoteUserHeader, resp.UserID)

		// finally, pass the request along the processing chain
		h.ServeHTTP(w, r)
	})
}
