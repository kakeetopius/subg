// Package opensubtitlesorg is used to get subtitles from opensubtitles.org
package opensubtitlesorg

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/kakeetopius/subg/internal/sessions"
	"github.com/kakeetopius/subg/internal/subformat"
	"github.com/kakeetopius/subg/internal/util"
	"github.com/kakeetopius/subg/internal/zip"
	"github.com/pterm/pterm"
)

const (
	domain            = "www.opensubtitles.org"
	downloadSubDomain = "dl.opensubtitles.org"

	openSubtitlesBaseURL = "https://" + domain
	opensubtitlesDLURL   = "https://" + downloadSubDomain

	searchPath = "/en/search"
	dlPath     = "/en/download/sub"

	anubisCookieName = "techaro.lol-anubis-auth"
	sessionTTL       = 5 * time.Hour
)

func (p *OpenSubtitlesOrg) search(opts Options) ([]OSOrgSubtitle, error) {
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
	p.session = session

	resp, err := p.session.DoRequest(encodeParams(searchPath, opts))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	title := getSubtitleTitle(doc)
	if title == "" {
		return nil, nil
	}

	subs := getSubtitleResults(doc)

	return subs, nil
}

func (p *OpenSubtitlesOrg) downloadSubtitle(subtitle *OSOrgSubtitle) (name string, subBytes io.ReadCloser, format subformat.FormatType, err error) {
	p.session.Client.WithBaseURL(opensubtitlesDLURL)

	spinner, err := pterm.DefaultSpinner.Start("Downloading Subtitle.........")
	if err != nil {
		return
	}
	defer func() {
		if err == nil {
			spinner.Success("Download Done")
		} else {
			spinner.Fail()
		}
	}()

	resp, err := p.session.DoRequest(pathFromID(subtitle.SubtitleID))
	if err != nil {
		return
	}
	defer resp.Body.Close()

	subFiles, err := zip.SubtitleFilesFromZip(resp.Body)
	if err != nil {
		return "", nil, 0, err
	}
	if len(subFiles) == 0 {
		err = fmt.Errorf("no subtitle files found in the zip file")
		return
	}

	subFile := subFiles[0]
	format, err = subformat.SubFormatFromFileName(subFile.Name)
	if err != nil {
		return "", nil, 0, err
	}
	subBytes, err = subFile.Open()
	if err != nil {
		return "", nil, 0, err
	}

	return util.StripExtension(subFile.Name), subBytes, format, nil
}

func pathFromID(s string) string {
	return dlPath + "/" + s
}

func extractSubID(path string) string {
	elem := strings.Split(path, "/")
	return elem[len(elem)-1]
}

func getSubtitleTitle(doc *goquery.Document) string {
	title := ""
	doc.Find(`#subtitles_body > div.content > div:nth-child(13) > div > h1`).Each(func(i int, s *goquery.Selection) {
		title = s.Text()
	})

	return title
}

func getSubtitleResults(doc *goquery.Document) []OSOrgSubtitle {
	subs := make([]OSOrgSubtitle, 0, 5)
	rows := doc.Find(`#search_results > tbody > tr`)

	rows.Each(func(i int, s *goquery.Selection) {
		if !isResult(s) {
			return
		}

		title := s.Find(`td:nth-child(1) a`).First().Text()
		lang, _ := s.Find(`td:nth-child(2) > a`).Attr("title")
		if lang != "English" {
			return
		}
		dlPath, _ := s.Find(`td:nth-child(5) > a`).Attr("href")
		votes := s.Find(`td:nth-child(6)`).First().Text()
		subs = append(subs, OSOrgSubtitle{
			Name:       cleanResultsTitle(title),
			SubtitleID: extractSubID(dlPath),
			Language:   lang,
			Votes:      votes,
		})
	})

	return subs
}

func isResult(row *goquery.Selection) bool {
	_, exists := row.Attr("onclick")
	return exists
}

func cleanResultsTitle(t string) string {
	words := strings.Split(t, "\n")
	for i := range words {
		words[i] = strings.TrimSpace(words[i])
	}

	return strings.Join(words, " ")
}

func encodeParams(baseURL string, opts Options) string {
	sb := strings.Builder{}

	filters := make([]string, 0, 10)

	if opts.Query != "" {
		filters = append(filters, fmt.Sprintf("%v-%v", "moviename", encodeName(opts.Query)))
	}

	if opts.IsSerie {
		filters = append(filters, fmt.Sprintf("%v-%v", "searchonlytvseries", "on"))
	}
	if opts.IsMovie {
		filters = append(filters, fmt.Sprintf("%v-%v", "searchonlymovies", "on"))
	}
	if opts.Season != 0 {
		filters = append(filters, fmt.Sprintf("%v-%v", "season", opts.Season))
	}
	if opts.Episode != 0 {
		filters = append(filters, fmt.Sprintf("%v-%v", "episode", opts.Episode))
	}
	if opts.Language != "" {
		filters = append(filters, fmt.Sprintf("%v-%v", "movielanguage", "english"))
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
	if len(words) == 0 {
		return ""
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
