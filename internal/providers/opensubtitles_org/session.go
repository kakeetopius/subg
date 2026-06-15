package opensubtitlesorg

import (
	"errors"
	"net/http"
	"time"
)

var ErrDomainSessionNotFound = errors.New("session for domain not found")

type Session struct {
	Domain         string `json:"domain"`
	CookieHeader   string `json:"cookie_header"`
	UserAgent      string `json:"user_agent"`
	AcquiredAtUnix int64  `json:"acquired_at_unix"`
}

func (s Session) isExpired() bool {
	return time.Now().Unix()-s.AcquiredAtUnix > int64(sessionTTL.Seconds())
}

func (s Session) isUsable() bool {
	return s.CookieHeader != "" && s.UserAgent != ""
}

func (s Session) doRequest(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", s.UserAgent)
	req.Header.Set("Cookie", s.CookieHeader)
	client := &http.Client{}
	return client.Do(req)
}
