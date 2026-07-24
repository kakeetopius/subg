package sessions

import (
	"context"
	"net/http"
	"time"

	"github.com/kakeetopius/subg/internal/httpclient"
)

type Session struct {
	Domain         string `json:"domain"`
	CookieHeader   string `json:"cookie_header"`
	UserAgent      string `json:"user_agent"`
	AcquiredAtUnix int64  `json:"acquired_at_unix"`

	Client *httpclient.Client `json:"-"`
}

// isExpired checks if the session has expired based on the provided sessionTTL.
func (s Session) isExpired(sessionTTL time.Duration) bool {
	return time.Now().Unix()-s.AcquiredAtUnix > int64(sessionTTL.Seconds())
}

// isUsable checks if the session has the necessary information to be used for making requests.
func (s Session) isUsable() bool {
	return s.CookieHeader != "" && s.UserAgent != ""
}

// initSessionHTTPClient initializes the HTTP client for the session with the provided baseURL, user agent, and cookie header.
func (s *Session) initSessionHTTPClient(baseURL string) {
	s.Client = httpclient.New().WithBaseURL(baseURL).WithUserAgent(s.UserAgent).WithCookie(s.CookieHeader)
}

// DoRequest performs an HTTP GET request to the specified path.
func (s Session) DoRequest(ctx context.Context, path string) (*http.Response, error) {
	return s.Client.Get(ctx, path, nil)
}

// DoRequestJSON performs an HTTP GET request to the specified path and decodes the JSON response into the provided target.
func (s Session) DoRequestJSON(ctx context.Context, path string, target any) error {
	return s.Client.GetJSON(ctx, path, nil, target)
}
