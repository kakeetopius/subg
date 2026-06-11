package opensubtitles

import (
	"fmt"

	"charm.land/bubbles/v2/table"
	"github.com/kakeetopius/subg/internal/providers"
	"github.com/kakeetopius/subg/internal/ui"
)

type OpenSubtitles struct {
	opts options

	subtitles []OSSubtitle
}

type options struct {
	providers.Options
	OSOptions
}

type OSOptions struct {
	APIKey   string
	CacheDir string
}

func (p *OpenSubtitles) Name() string {
	return "opensubtitles"
}

func (p *OpenSubtitles) WithOptions(opts providers.Options) {
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

func (p *OpenSubtitles) WithSpecificOptions(opts any) {
	if options, ok := opts.(OSOptions); ok {
		p.opts.OSOptions = options
	}
}

func (p *OpenSubtitles) SearchSubtitle() error {
	subs, err := searchSubtitle(p.opts)
	if err != nil {
		return err
	}
	if len(p.subtitles) < 1 {
		if p.opts.IsSerie {
			return fmt.Errorf("no results returned for %v Season %v Episode %v", p.opts.Query, p.opts.Season, p.opts.Episode)
		}
		return fmt.Errorf("no Results returned for %v", p.opts.Query)
	}
	p.subtitles = subs
	return nil
}

func (p *OpenSubtitles) Download(subs []providers.Subtitle) error {
	for _, sub := range subs {
		subtitle := sub.(OSSubtitle)
		err := downloadSubtitle(p.opts, &subtitle)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *OpenSubtitles) DownloadBest() error {
	return nil
}

func (p *OpenSubtitles) DisplaySelections() ([]providers.Subtitle, error) {
	if len(p.subtitles) < 1 {
		return nil, fmt.Errorf("opensubtitles did not return any results")
	}
	columns := []table.Column{
		{Title: "ID", Width: 8},
		{Title: "Name", Width: 72},
		{Title: "Lang", Width: 10},
		{Title: "Rating", Width: 10},
		{Title: "Votes", Width: 10},
	}

	rows := []table.Row{}
	for _, subtitle := range p.subtitles {
		rows = append(rows, []string{
			subtitle.SubtitleID,
			subtitle.Release,
			subtitle.Language,
			fmt.Sprintf("%v", subtitle.Ratings),
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

func (p *OpenSubtitles) subtitleByID(id string) (providers.Subtitle, error) {
	for _, sub := range p.subtitles {
		if sub.SubtitleID == id {
			return &sub, nil
		}
	}

	return nil, fmt.Errorf("subtitle with id %v not found in results", id)
}
