// Package opensubtitlesorg is used to get subtitles from opensubtitles.org
package opensubtitlesorg

import (
	"fmt"
	"io"
	"time"

	"github.com/kakeetopius/subg/internal/providers/sessions"
)

const (
	domain               = "www.opensubtitles.org"
	openSubtitlesBaseURL = "https://" + domain
	searchURL            = "/en/search"
	anubisCookieName     = "techaro.lol-anubis-auth"
	sessionTTL           = 5 * time.Hour
)

type Options struct {
	CacheDir string
}

func Search(opts Options) error {
	mustHaveCookies := []string{anubisCookieName, "cf_clearance"}

	sessionManager := sessions.
		NewSessionManager().
		WithCacheDir(opts.CacheDir).
		WithSessionTTL(sessionTTL).
		WithMustHaveCookies(domain, mustHaveCookies...).
		WithWaitDuration(2 * time.Minute)

	session, err := sessionManager.GetSession(domain, openSubtitlesBaseURL, false)
	if err != nil {
		return err
	}

	resp, err := session.DoRequest(searchURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Println(string(body))

	return nil
}
