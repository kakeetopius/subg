// Package sessions manages authenticated sessions for HTTP providers that require it. It handles session cookies, authentication tokens, and other state required
// to make authenticated requests to the provider.
package sessions

import (
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
	cachefile = "sessions.json"
)

var ErrDomainSessionNotFound = errors.New("session for domain not found")

type ErrCookieNotFound struct {
	CookieName string
}

func (e ErrCookieNotFound) Error() string {
	return fmt.Sprintf("could not get cookie \"%s\"", e.CookieName)
}

type SessionManager struct {
	// Directory where the sessions should be stored or are currently stored
	CacheDir string
	// After how long should saved sessions be considered expired and hence force refresh
	SessionTTL time.Duration
	// Cookies that must exist in session for domains
	MustHaveCookies map[string][]string
	// Time to wait for clearance cookies if any
	WaitDuration time.Duration
}

func NewSessionManager() *SessionManager {
	return &SessionManager{WaitDuration: time.Minute * 4}
}

func (m *SessionManager) WithCacheDir(cacheDir string) *SessionManager {
	m.CacheDir = cacheDir
	return m
}

func (m *SessionManager) WithSessionTTL(sessionTTL time.Duration) *SessionManager {
	m.SessionTTL = sessionTTL
	return m
}

func (m *SessionManager) WithWaitDuration(waitDuration time.Duration) *SessionManager {
	m.WaitDuration = waitDuration
	return m
}

func (m *SessionManager) WithMustHaveCookies(domain string, cookieNames ...string) *SessionManager {
	if m.MustHaveCookies == nil {
		m.MustHaveCookies = make(map[string][]string)
	}

	m.MustHaveCookies[domain] = cookieNames
	return m
}

// GetSession returns a session for the specified domain. challengeURL specifies the full URL of the page containing any challenge that must be completed to
// obtain valid session cookies. If challengeURL is empty, it defaults to "https://<domain>".
func (m SessionManager) GetSession(domain, challengeURL string, forceRefresh bool) (Session, error) {
	baseURL := fmt.Sprintf("https://%v", domain)
	if challengeURL == "" {
		challengeURL = baseURL
	}

	if !forceRefresh {
		cachedSession, err := m.loadSessionFromCachedFile(domain)
		if err == nil {
			if !cachedSession.isExpired(m.SessionTTL) && cachedSession.isUsable() {
				cachedSession.initSessionHTTPClient(baseURL)
				return cachedSession, err
			}
		} else if !errors.Is(err, fs.ErrNotExist) && !errors.Is(err, ErrDomainSessionNotFound) {
			return Session{}, err
		}
	}

	s, err := m.acquireSession(domain, challengeURL)
	if err != nil {
		return Session{}, err
	}
	if err := m.saveSessionToCacheFile(domain, s); err != nil {
		return Session{}, err
	}

	s.initSessionHTTPClient(baseURL)
	return s, nil
}

// loadSessionFromCachedFile loads the session for the given domain from the cache file. If the session is not found, it returns ErrDomainSessionNotFound.
func (m SessionManager) loadSessionFromCachedFile(domain string) (Session, error) {
	path := filepath.Join(m.CacheDir, cachefile)

	data, err := os.ReadFile(path)
	if err != nil {
		return Session{}, err
	}

	sessions := make(map[string]Session)
	if err := json.Unmarshal(data, &sessions); err != nil {
		return Session{}, err
	}
	s, ok := sessions[domain]
	if !ok {
		return Session{}, ErrDomainSessionNotFound
	}
	return s, nil
}

// saveSessionToCacheFile saves the session for the given domain to the cache file. If the session for the domain already exists in the cache file, it will be overwritten.
func (m SessionManager) saveSessionToCacheFile(domain string, s Session) error {
	path := filepath.Join(m.CacheDir, cachefile)

	err := util.CreateDirIfNotExists(filepath.Dir(path))
	if err != nil {
		return err
	}

	sessions := make(map[string]Session)

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

// acquireSession launches a browser and navigates to the challenge URL, waiting for the required session cookies to be set. It returns a Session object containing the acquired cookies and user agent.
func (m SessionManager) acquireSession(domain, challengeURL string) (Session, error) {
	fmt.Println("\nBrowser window opened. Complete any challenges provided if prompted.")
	fmt.Printf("Waiting up to %s for clearance cookies...\n", m.WaitDuration.String())

	browserCtx, cancel := m.startBrowser()
	defer cancel()

	tabCtx, cancel, err := m.goToSiteWithBrowser(browserCtx, challengeURL)
	if err != nil {
		return Session{}, err
	}
	defer cancel()

	deadline := time.Now().Add(m.WaitDuration)

	var cookies []*network.Cookie

	var validadtionErr error
	for time.Now().Before(deadline) {
		cookies, err = m.fetchCookies(tabCtx)
		if err != nil {
			return Session{}, fmt.Errorf("error getting cookies: %w", err)
		}

		validadtionErr = m.validateCookies(cookies, domain)
		if validadtionErr == nil {
			// if all cookies were got before deadline
			break
		}

		if _, ok := errors.AsType[ErrCookieNotFound](validadtionErr); !ok {
			// if error is anything other than cookie not found, return it.
			return Session{}, err
		}

		time.Sleep(1 * time.Second)
	}

	if validadtionErr != nil {
		return Session{}, validadtionErr
	}

	// Build cookie header from ALL cookies for the domain
	cookieHeader := m.buildCookieString(cookies, domain)
	if cookieHeader == "" {
		return Session{}, fmt.Errorf("no cookies found for domain")
	}

	ua, err := m.getUserAgent(tabCtx)
	if err != nil {
		return Session{}, fmt.Errorf("error getting user agent: %w", err)
	}

	return Session{
		Domain:         domain,
		CookieHeader:   cookieHeader,
		UserAgent:      ua,
		AcquiredAtUnix: time.Now().Unix(),
	}, nil
}

// startBrowser starts a new Chrome browser instance. It returns a context and a cancel function that can be used to stop the browser.
func (m SessionManager) startBrowser() (context.Context, context.CancelFunc) {
	opts := append(
		chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.WindowSize(1280, 900),
	)

	return chromedp.NewExecAllocator(context.Background(), opts...)
}

// goToSiteWithBrowser navigates to the specified URL using the provided browser context. It returns a new context for the tab, a cancel function to close the tab, and an error if navigation fails.
func (m SessionManager) goToSiteWithBrowser(browserCtx context.Context, url string) (context.Context, context.CancelFunc, error) {
	ctx, cancel := chromedp.NewContext(browserCtx)

	err := chromedp.Run(ctx, chromedp.Navigate(url))
	if err != nil {
		return nil, nil, fmt.Errorf("error navigating to website: %w", err)
	}

	return ctx, cancel, nil
}

// validateCookies checks if the required cookies for the given domain are present in the provided list of cookies. If any required cookie is missing, it returns an ErrCookieNotFound error.
func (m SessionManager) validateCookies(cookies []*network.Cookie, domain string) error {
	mustHave, found := m.MustHaveCookies[domain]
	if !found {
		return nil
	}

	domainCookiesFound := make(map[string]struct{}, len(cookies))
	for _, c := range cookies {
		if !domainMatches(c.Domain, domain) {
			continue
		}
		domainCookiesFound[c.Name] = struct{}{}
	}

	for _, cookieName := range mustHave {
		_, ok := domainCookiesFound[cookieName]
		if !ok {
			return ErrCookieNotFound{CookieName: cookieName}
		}
	}

	return nil
}

// fetchCookies retrieves all cookies from the current browser/tab context.
func (m SessionManager) fetchCookies(ctx context.Context) (cookies []*network.Cookie, err error) {
	err = chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		cookies, err = network.GetCookies().Do(ctx)
		return err
	}))

	return
}

// buildCookieString constructs a cookie header string from the provided list of cookies, filtering by the specified domain. It returns a string in the format "name1=value1; name2=value2; ...".
func (m SessionManager) buildCookieString(cookies []*network.Cookie, domain string) string {
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

// getUserAgent retrieves the user agent string from the current browser/tab context. It returns the user agent string and an error if retrieval fails.
func (m SessionManager) getUserAgent(ctx context.Context) (string, error) {
	var ua string
	err := chromedp.Run(ctx, chromedp.Evaluate(`navigator.userAgent`, &ua))
	return ua, err
}

// domainMatches checks if the cookie domain matches the target domain. It returns true if the cookie domain is equal to the target domain or if the cookie domain is a subdomain of the target domain.
func domainMatches(cookieDomain, target string) bool {
	cd := strings.TrimPrefix(strings.ToLower(cookieDomain), ".")
	t := strings.ToLower(target)
	return cd == t || strings.HasSuffix(cd, t)
}
