package opensubtitlesorg

import (
	"fmt"

	"charm.land/bubbles/v2/table"
	"github.com/kakeetopius/subg/internal/providers"
	"github.com/kakeetopius/subg/internal/ui"
)

type OpenSubtitlesOrg struct {
	opts Options

	subtitles []OSSubtitle
}

type OSSubtitle struct {
	Name        string
	SubtitleID  string
	Release     string
	Votes       string
	DownloadURL string
	Language    string
}
type Options struct {
	providers.Options
	OSOptions
}

type OSOptions struct {
	CacheDir string
}

func (p *OpenSubtitlesOrg) Name() string {
	return "opensubtitles.org"
}

func (p *OpenSubtitlesOrg) WithOptions(opts providers.Options) {
	featureType := "all"
	switch {
	case opts.IsMovie:
		featureType = "movie"
	case opts.IsSerie || opts.Season != 0 || opts.Episode != 0:
		// if a season or episode is given we assume it is a serie
		featureType = "episode"
	}

	opts.Type = featureType
	p.opts.Options = opts
}

func (p *OpenSubtitlesOrg) WithSpecificOptions(opts any) {
	if options, ok := opts.(OSOptions); ok {
		p.opts.OSOptions = options
	}
}

func (p *OpenSubtitlesOrg) SearchSubtitle() error {
	subs, err := Search(p.opts)
	if err != nil {
		return err
	}
	if len(subs) == 0 {
		if p.opts.IsSerie {
			return fmt.Errorf("no results returned for %v Season %v Episode %v", p.opts.Query, p.opts.Season, p.opts.Episode)
		}
		return fmt.Errorf("no Results returned for %v", p.opts.Query)
	}
	p.subtitles = subs
	return nil
}

func (p *OpenSubtitlesOrg) Download(subs []providers.Subtitle) error {
	for _, sub := range subs {
		subtitle := sub.(OSSubtitle)
		err := downloadSubtitle(p.opts, &subtitle)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *OpenSubtitlesOrg) DownloadBest() error {
	return nil
}

func (p *OpenSubtitlesOrg) DisplaySelections() ([]providers.Subtitle, error) {
	if len(p.subtitles) < 1 {
		return nil, fmt.Errorf("opensubtitles did not return any results")
	}
	columns := []table.Column{
		{Title: "ID", Width: 8},
		{Title: "Name", Width: 72},
		{Title: "Lang", Width: 10},
		{Title: "Votes", Width: 10},
	}

	rows := []table.Row{}
	for _, subtitle := range p.subtitles {
		rows = append(rows, []string{
			subtitle.SubtitleID,
			subtitle.Release,
			subtitle.Language,
			fmt.Sprintf("%v", subtitle.Votes),
		})
	}

	subID, err := ui.DisplayTableAndGetSubtitleID(rows, columns)
	if err != nil {
		return nil, err
	}
	sub, err := p.subtitleByID(subID)
	if err != nil {
		return nil, err
	}

	return []providers.Subtitle{sub}, nil
}

func (p *OpenSubtitlesOrg) subtitleByID(id string) (providers.Subtitle, error) {
	for _, sub := range p.subtitles {
		if sub.SubtitleID == id {
			return sub, nil
		}
	}

	return nil, fmt.Errorf("subtitle with id %v not found in results", id)
}
