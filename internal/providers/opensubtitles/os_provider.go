package opensubtitles

import (
	"context"
	"fmt"

	"charm.land/bubbles/v2/table"
	"github.com/kakeetopius/subg/internal/providers"
	"github.com/kakeetopius/subg/internal/ui"
)

type OpenSubtitles struct {
	OSOptions
	subtitles []OSSubtitle
}

type OSOptions struct {
	APIKey   string
	CacheDir string
}

func (OSSubtitle) IsSub() {}

func NewProvider(opts OSOptions) *OpenSubtitles {
	return &OpenSubtitles{
		OSOptions: opts,
	}
}

func (p *OpenSubtitles) Name() string {
	return "opensubtitles"
}

func (p *OpenSubtitles) SearchSubtitle(ctx context.Context, opts providers.Options) error {
	featureType := "all"
	switch {
	case opts.IsMovie:
		featureType = "movie"
	case opts.IsSerie || opts.Season != 0 || opts.Episode != 0:
		// if a season or episode is given we assume it is a serie
		featureType = "episode"
	}
	opts.Type = featureType

	subs, err := p.searchSubtitle(opts)
	if err != nil {
		return err
	}
	if len(subs) == 0 {
		if opts.IsSerie {
			return fmt.Errorf("no results returned for %v Season %v Episode %v", opts.Query, opts.Season, opts.Episode)
		}
		return fmt.Errorf("no Results returned for %v", opts.Query)
	}
	p.subtitles = subs
	return nil
}

func (p *OpenSubtitles) Download(ctx context.Context, sub providers.Subtitle) (providers.SubtitleFile, error) {
	subtitle := sub.(OSSubtitle)
	return p.downloadSubtitle(ctx, &subtitle)
}

func (p *OpenSubtitles) DownloadBest(ctx context.Context) (providers.SubtitleFile, error) {
	return p.downloadSubtitle(ctx, &p.subtitles[0])
}

func (p *OpenSubtitles) PromptSelection() ([]providers.Subtitle, error) {
	if len(p.subtitles) < 1 {
		return nil, fmt.Errorf("opensubtitles did not return any results")
	}
	maxNameWidth := 72
	columns := []table.Column{
		{Title: "ID", Width: 8},
		{Title: "Name", Width: maxNameWidth},
		{Title: "Lang", Width: 10},
		{Title: "Rating", Width: 10},
		{Title: "Votes", Width: 10},
	}

	rows := []table.Row{}
	maxNameLen := 0
	for _, subtitle := range p.subtitles {
		if len(subtitle.Release) > maxNameLen {
			maxNameLen = len(subtitle.Release)
		}
		rows = append(rows, []string{
			subtitle.SubtitleID,
			subtitle.Release,
			subtitle.Language,
			fmt.Sprintf("%v", subtitle.Ratings),
			fmt.Sprintf("%v", subtitle.Votes),
		})
	}

	if maxNameLen < maxNameWidth {
		columns[1].Width = maxNameLen
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
			return sub, nil
		}
	}

	return nil, fmt.Errorf("subtitle with id %v not found in results", id)
}
