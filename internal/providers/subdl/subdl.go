package subdl

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/kakeetopius/subg/internal/providers"
	"github.com/kakeetopius/subg/internal/sessions"
	"github.com/kakeetopius/subg/internal/subformat"
	"github.com/kakeetopius/subg/internal/util"
	"github.com/kakeetopius/subg/internal/zip"
	"github.com/pterm/pterm"
)

const (
	domain          = "subdl.com"
	searchSubDomain = "api3.subdl.com"
	subdlBaseURL    = "https://" + domain
	subdlSearchURL  = "https://" + searchSubDomain
)

func (p *SubDL) searchSubtitles(ctx context.Context, opts providers.SearchOptions) ([]providers.Subtitle, error) {
	sessionManager := sessions.NewSessionManager()

	session, err := sessionManager.GetSession(domain, subdlBaseURL, false)
	if err != nil {
		return nil, err
	}
	p.session = session

	spinner, err := pterm.DefaultSpinner.Start("Searching subtitles on subdl.com..........")
	defer func() {
		if err == nil {
			spinner.Success("Search Complete")
		} else {
			spinner.Fail()
		}
	}()

	doc, err := p.getSubtitlesPage(ctx, &opts)
	if err != nil {
		return nil, err
	}

	return p.parseSubtitlePage(doc, &opts)
}

func (p *SubDL) downloadSubtitle(ctx context.Context, sub SDSubtitle) (subtitleFile providers.SubtitleFile, err error) {
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

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sub.DownloadLink, nil)
	if err != nil {
		return
	}
	resp, err := p.session.Do(req)
	if err != nil {
		return providers.SubtitleFile{}, err
	}
	defer resp.Body.Close()

	subFiles, err := zip.SubtitleFilesFromZip(resp.Body)
	if err != nil {
		return
	}
	if len(subFiles) == 0 {
		err = fmt.Errorf("no subtitle files found in the downloaded zip file")
		return
	}

	subFile := subFiles[0]
	format, err := subformat.SubFormatFromFileName(subFile.Name)
	if err != nil {
		return
	}
	subBytes, err := subFile.Open()
	if err != nil {
		return
	}

	return providers.SubtitleFile{
		Name:       util.StripExtension(subFile.Name),
		Type:       format,
		ReadCloser: subBytes,
	}, nil
}

func (p *SubDL) parseSubtitlePage(doc *goquery.Document, opts *providers.SearchOptions) ([]providers.Subtitle, error) {
	subsLi := doc.Find("ul").Find("li")

	subs := make([]providers.Subtitle, 0, 5)

	subsLi.Each(func(i int, s *goquery.Selection) {
		if !isSubtitleRow(s) {
			return
		}

		if opts.Type == "tv" && !episodeMatches(s, opts.Episode) {
			return
		}

		link := getSubtitleDownloadLink(s)
		name := getSubtitleName(s)
		author := getSubtitleAuthor(s)
		id := getSubtitleID(s)

		subs = append(subs, SDSubtitle{
			Name:         name,
			SubID:        id,
			Author:       author,
			DownloadLink: link,
		})
	})

	return subs, nil
}

func (p *SubDL) getSubtitlesPage(ctx context.Context, opts *providers.SearchOptions) (*goquery.Document, error) {
	searchResult, err := p.getSearchResult(ctx, opts)
	if err != nil {
		return nil, err
	}

	p.session.WithBaseURL(subdlBaseURL)
	searchResult.Link += "/english"

	// for serie returns a page listing the seasons for movie returns the page with the subtitles in english
	resp, err := p.session.Get(ctx, searchResult.Link, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	if opts.Type == "movie" {
		return doc, nil
	}

	links, err := p.getSeasonLinks(doc, opts)
	if err != nil {
		return nil, err
	}

	p.session.WithBaseURL(subdlBaseURL)
	if _, found := links[opts.Season]; !found {
		return nil, providers.ErrNoResultsFound{Query: opts.Query, Episode: opts.Episode, Season: opts.Season}
	}

	// for serie, returns the page with the subtitles for the season in english
	resp, err = p.session.Get(ctx, links[opts.Season]+"/english", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return goquery.NewDocumentFromReader(resp.Body)
}

func (p *SubDL) getSearchResult(ctx context.Context, opts *providers.SearchOptions) (SubdlSearchResult, error) {
	p.session.WithBaseURL(subdlSearchURL)

	var searchResults SubdlSearchResults

	err := p.session.GetJSON(ctx, "/auto", SubDlSearchParams{Query: opts.Query}, &searchResults)
	if err != nil {
		return SubdlSearchResult{}, err
	}

	if len(searchResults.Results) == 0 {
		return SubdlSearchResult{}, providers.ErrNoResultsFound{Query: opts.Query, Episode: opts.Episode, Season: opts.Season}
	}

	for _, res := range searchResults.Results {
		if res.Type != opts.Type {
			continue
		}
		if opts.Year != 0 && opts.Year != res.Year {
			continue
		}
		return res, nil
	}

	return SubdlSearchResult{}, providers.ErrNoResultsFound{Query: opts.Query, Episode: opts.Episode, Season: opts.Season}
}

func (p *SubDL) getSeasonLinks(doc *goquery.Document, opts *providers.SearchOptions) (map[int]string, error) {
	seasonLinks := make(map[int]string)

	links := doc.Find("body > div.wrapper.pb-16 > div > div.min-h-\\[150px\\] > div > div.mt-5.flex.flex-col.gap-4 > a")

	links.Each(func(i int, s *goquery.Selection) {
		link, ok := s.Attr("href")
		if !ok {
			return
		}
		// link is of the form /subtitle/sd*****/manifest/first-season
		parts := strings.Split(link, "/")
		seasonNumber := mapOrdinalToNumber(strings.Split(parts[len(parts)-1], "-")[0])

		if seasonNumber == 0 {
			return
		}

		seasonLinks[seasonNumber] = link
	})

	if _, ok := seasonLinks[opts.Season]; !ok {
		return nil, providers.ErrNoResultsFound{Query: opts.Query, Episode: opts.Episode, Season: opts.Season}
	}

	return seasonLinks, nil
}

func isSubtitleRow(s *goquery.Selection) bool {
	_, exists := s.Attr("data-id")
	return exists
}

func episodeMatches(s *goquery.Selection, episode int) bool {
	episodeFromStr, _ := s.Attr("data-episode-from")
	episodeToStr, _ := s.Attr("data-episode-to")

	episodeFrom, _ := strconv.Atoi(episodeFromStr)
	episodeTo, _ := strconv.Atoi(episodeToStr)

	return episode == episodeFrom && episode == episodeTo
}

func getSubtitleDownloadLink(s *goquery.Selection) string {
	link, _ := s.Find("div").Eq(1).Find("a").First().Attr("href")
	return link
}

func getSubtitleName(s *goquery.Selection) string {
	return cleanSubtitleName(s.Find("div").First().Find("a > h4").Text())
}

func getSubtitleID(s *goquery.Selection) string {
	id, _ := s.Attr("data-id")
	return id
}

func getSubtitleAuthor(s *goquery.Selection) string {
	// author name is in the form "(author)"
	name := s.Find("div").First().Find("a").Eq(1).Text()

	if name == "" {
		return name
	}

	firstBracket := strings.Index(name, "(")
	secondBracket := strings.Index(name, ")")
	if firstBracket == -1 || secondBracket == -1 {
		return ""
	}
	return name[firstBracket+1 : secondBracket]
}

func cleanSubtitleName(t string) string {
	words := strings.Split(t, "\n")
	for i := range words {
		words[i] = strings.TrimSpace(words[i])
	}

	return strings.Join(words, " ")
}
