// Package opensubtitlesorg is used to get subtitles from opensubtitles.org
package opensubtitlesorg

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/kakeetopius/subg/internal/util"
)

const (
	domain               = "www.opensubtitles.org"
	openSubtitlesBaseURL = "https://" + domain
	searchURL            = openSubtitlesBaseURL + "/en/search"
	anubisCookieName     = "techaro.lol-anubis-auth"
	sessionTTL           = 5 * time.Hour
	cachefile            = "sessions.json"
)

type Options struct {
	CacheDir string
}

func Search(opts Options) error {
	session, err := EnsureSession(domain, searchURL, opts.CacheDir, false)
	if err != nil {
		return err
	}

	fmt.Printf("Got session: cf=%v anubis=%v\n", session.CfClearance, session.AnubisAuth)

	resp, err := session.doRequest(searchURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func EnsureSession(domain, challengeURL, cacheDir string, forceRefresh bool) (Session, error) {
	if !forceRefresh {
		cachedSession, err := loadSessionFromCachedFile(domain, cacheDir)
		if err == nil {
			if !cachedSession.isExpired() && cachedSession.isUsable() {
				return *cachedSession, err
			}
		} else if !errors.Is(err, fs.ErrNotExist) && !errors.Is(err, ErrDomainSessionNotFound) {
			return Session{}, err
		}
	}

	s, err := acquireSession(domain, challengeURL)
	if err != nil {
		return Session{}, err
	}
	if err := saveSessionToCacheFile(domain, s, cacheDir); err != nil {
		return Session{}, err
	}
	return s, nil
}

func loadSessionFromCachedFile(domain string, cacheDir string) (*Session, error) {
	path := filepath.Join(cacheDir, cachefile)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var sessions map[string]Session
	if err := json.Unmarshal(data, &sessions); err != nil {
		return nil, err
	}
	s, ok := sessions[domain]
	if !ok {
		return nil, ErrDomainSessionNotFound
	}
	return &s, nil
}

func saveSessionToCacheFile(domain string, s Session, cacheDir string) error {
	path := filepath.Join(cacheDir, cachefile)

	err := util.CreateDirIfNotExists(filepath.Dir(path))
	if err != nil {
		return err
	}

	sessions := map[string]Session{}
	data, err := os.ReadFile(path)
	if err == nil {
		_ = json.Unmarshal(data, &sessions)
	}
	sessions[domain] = s

	out, err := json.MarshalIndent(sessions, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, out, 0o644)
}

// acquireSession launches a real, head-full Chrome to clear both the
// Cloudflare challenge (cf_clearance) and the Anubis PoW challenge
// (within.website-x-cmd-anubis-auth), then extracts cookies + UA.
func acquireSession(domain, challengeURL string) (Session, error) {
	fmt.Println("Browser window opened. Complete any Cloudflare/Anubis challenge if prompted.")
	fmt.Println("Waiting up to 4 minutes for clearance cookies...")

	browserCtx, cancel := startBrowser()
	defer cancel()

	tabCtx, cancel, err := goToSiteWithBrowser(browserCtx, challengeURL)
	if err != nil {
		return Session{}, err
	}
	defer cancel()

	deadline := time.Now().Add(4 * time.Minute)

	var cookies []*network.Cookie
	var cfValue, anubisValue string

	for time.Now().Before(deadline) {
		cookies, err = fetchCookies(tabCtx)
		if err != nil {
			return Session{}, fmt.Errorf("error getting cookies: %w", err)
		}

		cfValue, anubisValue = getAuthCookies(cookies)

		// Done once we have whichever cookies the site actually issues.
		// At minimum we need *some* indication a challenge was passed:
		// either cookie being present is sufficient progress.
		if cmp.Or(cfValue, anubisValue) != "" {
			// Give the page a brief moment in case both challenges
			// resolve in sequence (Cloudflare first, then Anubis).
			time.Sleep(2 * time.Second)

			// Re-fetch once more to catch a cookie that appears just after.
			cookies, err = fetchCookies(tabCtx)
			if err != nil {
				return Session{}, fmt.Errorf("get cookies (recheck): %w", err)
			}

			cfValue, anubisValue = getAuthCookies(cookies)
			break
		}

		time.Sleep(1 * time.Second)
	}

	if cmp.Or(cfValue, anubisValue) == "" {
		return Session{}, fmt.Errorf("timed out waiting for cf_clearance / anubis auth cookie")
	}

	// Build cookie header from ALL cookies for the domain
	cookieHeader := buildCookieString(cookies)
	if cookieHeader == "" {
		return Session{}, fmt.Errorf("no cookies found for domain")
	}

	ua, err := getUserAgent(tabCtx)
	if err != nil {
		return Session{}, fmt.Errorf("error getting user agent: %w", err)
	}

	return Session{
		Domain:         domain,
		CookieHeader:   cookieHeader,
		CfClearance:    cfValue,
		AnubisAuth:     anubisValue,
		UserAgent:      ua,
		AcquiredAtUnix: time.Now().Unix(),
	}, nil
}

func startBrowser() (context.Context, context.CancelFunc) {
	opts := append(
		chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.WindowSize(1280, 900),
	)

	return chromedp.NewExecAllocator(context.Background(), opts...)
}

func goToSiteWithBrowser(browserCtx context.Context, url string) (context.Context, context.CancelFunc, error) {
	ctx, cancel := chromedp.NewContext(browserCtx)

	err := chromedp.Run(ctx, chromedp.Navigate(url))
	if err != nil {
		return nil, nil, fmt.Errorf("error navigating to website: %w", err)
	}

	return ctx, cancel, nil
}

func getAuthCookies(cookies []*network.Cookie) (cfValue string, anubisValue string) {
	for _, c := range cookies {
		if !domainMatches(c.Domain, domain) {
			continue
		}
		switch c.Name {
		case "cf_clearance":
			cfValue = c.Value
		case anubisCookieName:
			anubisValue = c.Value
		}
	}

	return
}

func fetchCookies(ctx context.Context) (cookies []*network.Cookie, err error) {
	err = chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		cookies, err = network.GetCookies().Do(ctx)
		return err
	}))

	return
}

func buildCookieString(cookies []*network.Cookie) string {
	var sb strings.Builder
	for _, c := range cookies {
		if !domainMatches(c.Domain, domain) {
			continue
		}
		if sb.Len() > 0 {
			sb.WriteString("; ")
		}
		sb.WriteString(c.Name)
		sb.WriteString("=")
		sb.WriteString(c.Value)
	}
	return sb.String()
}

func getUserAgent(ctx context.Context) (string, error) {
	var ua string
	err := chromedp.Run(ctx, chromedp.Evaluate(`navigator.userAgent`, &ua))
	return ua, err
}

func domainMatches(cookieDomain, target string) bool {
	cd := strings.TrimPrefix(strings.ToLower(cookieDomain), ".")
	t := strings.ToLower(target)
	return cd == t || strings.HasSuffix(t, cd)
}
