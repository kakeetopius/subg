package subdl

import (
	"fmt"
	"io"

	"charm.land/bubbles/v2/table"
	"github.com/kakeetopius/subg/internal/formats"
	"github.com/kakeetopius/subg/internal/providers"
	"github.com/kakeetopius/subg/internal/ui"
)

type SubDL struct {
	opts options

	subtitles []SDSubtitle
}

type options struct {
	providers.Options
	SubDLOpts
}

type SubDLOpts struct {
	APIKey string
}

func NewProvider(opts SubDLOpts) *SubDL {
	return &SubDL{
		opts: options{
			SubDLOpts: opts,
		},
	}
}

func (p *SubDL) Name() string {
	return "subdl.com"
}

func (p *SubDL) WithOptions(opts providers.Options) {
	// keyword "movie" or "tv" is what is required by the subdl API.
	featureType := "movie"
	if p.opts.IsSerie || p.opts.Episode != 0 || p.opts.Season != 0 {
		// if a season or episode is given we assume it is a serie
		featureType = "tv"
	}
	p.opts.Type = featureType

	p.opts.Options = opts
}

func (p *SubDL) SearchSubtitle() error {
	subs, err := searchSubtitles(p.opts)
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

func (p *SubDL) Download(sub providers.Subtitle) (string, io.ReadCloser, formats.FormatType, error) {
	subtitle := sub.(SDSubtitle)
	return downloadSubtitle(&subtitle)
}

func (p *SubDL) DownloadBest() error {
	return nil
}

func (p *SubDL) DisplaySelections() ([]providers.Subtitle, error) {
	if len(p.subtitles) == 0 {
		return nil, fmt.Errorf("no subtitles returned by subdl")
	}

	maxNameWidth := 70
	columns := []table.Column{
		{Title: "ID", Width: 5},
		{Title: "Name", Width: maxNameWidth},
		{Title: "Lang", Width: 10},
		{Title: "Author", Width: 15},
	}

	rows := []table.Row{}
	maxNameLen := 0
	for _, sub := range p.subtitles {
		if len(sub.ReleaseName) > maxNameLen {
			maxNameLen = len(sub.ReleaseName)
		}
		rows = append(rows, []string{
			fmt.Sprint(sub.ID),
			sub.ReleaseName,
			sub.Lang,
			sub.Author,
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

func (p *SubDL) subtitleByID(id string) (providers.Subtitle, error) {
	for _, sub := range p.subtitles {
		idStr := fmt.Sprint(sub.ID)
		if idStr == id {
			return sub, nil
		}
	}

	return nil, fmt.Errorf("subtitle with id %v not found in results", id)
}
