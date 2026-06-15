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

	client *httpclient.Client
}

func (s Session) isExpired(sessionTTL time.Duration) bool {
	return time.Now().Unix()-s.AcquiredAtUnix > int64(sessionTTL.Seconds())
}

func (s Session) isUsable() bool {
	return s.CookieHeader != "" && s.UserAgent != ""
}

func (s *Session) initSessionHTTPClient(baseURL string) {
	s.client = httpclient.New().WithBaseURL(baseURL).WithUserAgent(s.UserAgent).WithCookie(s.CookieHeader)
}

func (s Session) DoRequest(path string) (*http.Response, error) {
	return s.client.Get(context.Background(), path, nil)
}
