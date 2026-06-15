// Package opensubtitlesorg is used to get subtitles from opensubtitles.org
package opensubtitlesorg

import (
	"fmt"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/kakeetopius/subg/internal/providers/sessions"
)

const (
	domain               = "www.opensubtitles.org"
	openSubtitlesBaseURL = "https://" + domain
	searchURL            = "/en/search"
	anubisCookieName     = "techaro.lol-anubis-auth"
	sessionTTL           = 5 * time.Hour
)

func Search(opts Options) ([]OSSubtitle, error) {
	mustHaveCookies := []string{anubisCookieName, "cf_clearance"}

	sessionManager := sessions.
		NewSessionManager().
		WithCacheDir(opts.CacheDir).
		WithSessionTTL(sessionTTL).
		WithMustHaveCookies(domain, mustHaveCookies...).
		WithWaitDuration(2 * time.Minute)

	session, err := sessionManager.GetSession(domain, openSubtitlesBaseURL, false)
	if err != nil {
		return nil, err
	}

	resp, err := session.DoRequest(encodeParams(searchURL, opts))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	title := getSubtitleTitle(doc)
	fmt.Println("Title: ", title)

	fmt.Println("Results: ", getSubtitleResults(doc))

	return nil, nil
}

func downloadSubtitle(opts Options, subtitle *OSSubtitle) (err error) {
	return nil
}

func getSubtitleTitle(doc *goquery.Document) string {
	title := ""
	doc.Find(`#subtitles_body > div.content > div:nth-child(13) > div > h1`).Each(func(i int, s *goquery.Selection) {
		title = s.Text()
	})

	return title
}

func getSubtitleResults(doc *goquery.Document) []OSSubtitle {
	subs := make([]OSSubtitle, 0, 5)
	rows := doc.Find(`#search_results > tbody > tr`)

	rows.Each(func(i int, s *goquery.Selection) {
		if i == 0 {
			return
		}
		title := s.Find(`td:nth-child(1) a`).First().Text()
		lang, _ := s.Find(`td:nth-child(2) > a`).Attr("title")
		dl, _ := s.Find(`td:nth-child(5) > a`).Attr("href")
		votes := s.Find(`td:nth-child(6)`).First().Text()
		subs = append(subs, OSSubtitle{
			Name:        title,
			Language:    lang,
			DownloadURL: dl,
			Votes:       votes,
		})
	})

	return subs
}

func encodeParams(baseURL string, opts Options) string {
	sb := strings.Builder{}

	filters := make([]string, 0, 10)

	if opts.Query != "" {
		filters = append(filters, fmt.Sprintf("%v-%v", "moviename", encodeName(opts.Query)))
	}

	filters = append(filters, fmt.Sprintf("%v-%v", "sublanguageid", "all"))
	if opts.Season != 0 {
		filters = append(filters, fmt.Sprintf("%v-%v", "season", opts.Season))
	}
	if opts.Episode != 0 {
		filters = append(filters, fmt.Sprintf("%v-%v", "episode", opts.Episode))
	}
	if opts.Language != "" {
		filters = append(filters, fmt.Sprintf("%v-%v", "movielanguage", "english"))
	}
	if opts.IsSerie {
		filters = append(filters, fmt.Sprintf("%v-%v", "searchonlytvseries", "on"))
	}
	if opts.IsSerie {
		filters = append(filters, fmt.Sprintf("%v-%v", "searchonlymovies", "on"))
	}
	if opts.Year != 0 {
		filters = append(filters, fmt.Sprintf("%v-%v", "movieyear", opts.Year))
		filters = append(filters, fmt.Sprintf("%v-%v", "movieyearsign", 1))
	}

	sb.WriteString(baseURL)
	for _, filter := range filters {
		sb.WriteString("/")
		sb.WriteString(filter)
	}

	return sb.String()
}

func encodeName(n string) string {
	words := strings.Split(n, " ")
	if len(words) == 1 {
		return words[0]
	}
	sb := strings.Builder{}

	for i, word := range words {
		if i != 0 {
			sb.WriteString("+")
		}
		sb.WriteString(word)

	}

	return sb.String()
}
