package opensubtitlesorg

import (
	"fmt"
	"io"

	"charm.land/bubbles/v2/table"
	"github.com/kakeetopius/subg/internal/providers"
	"github.com/kakeetopius/subg/internal/sessions"
	"github.com/kakeetopius/subg/internal/subformat"
	"github.com/kakeetopius/subg/internal/ui"
)

type OpenSubtitlesOrg struct {
	opts Options

	subtitles []OSOrgSubtitle

	session sessions.Session
}

type OSOrgSubtitle struct {
	Name       string
	SubtitleID string
	Release    string
	Votes      string
	Language   string
}

type Options struct {
	providers.Options
	OSOrgOptions
}

type OSOrgOptions struct {
	CacheDir string
}

func NewProvider(opts OSOrgOptions) *OpenSubtitlesOrg {
	return &OpenSubtitlesOrg{
		opts: Options{
			OSOrgOptions: opts,
		},
	}
}

func (p *OpenSubtitlesOrg) Name() string {
	return "opensubtitles.org"
}

func (p *OpenSubtitlesOrg) WithOptions(opts providers.Options) {
	if opts.Season != 0 || opts.Episode != 0 {
		opts.IsSerie = true
	} else {
		opts.IsMovie = true
	}

	p.opts.Options = opts
}

func (p *OpenSubtitlesOrg) SearchSubtitle() error {
	subs, err := p.search(p.opts)
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

func (p *OpenSubtitlesOrg) Download(sub providers.Subtitle) (string, io.ReadCloser, subformat.FormatType, error) {
	subtitle := sub.(OSOrgSubtitle)
	return p.downloadSubtitle(&subtitle)
}

func (p *OpenSubtitlesOrg) DownloadBest() (string, io.ReadCloser, subformat.FormatType, error) {
	return p.downloadSubtitle(&p.subtitles[0])
}

func (p *OpenSubtitlesOrg) DisplaySelections() ([]providers.Subtitle, error) {
	if len(p.subtitles) < 1 {
		return nil, fmt.Errorf("opensubtitles did not return any results")
	}
	maxNameWidth := 52
	columns := []table.Column{
		{Title: "ID", Width: 8},
		{Title: "Name", Width: maxNameWidth},
		{Title: "Lang", Width: 10},
		{Title: "Votes", Width: 12},
	}

	rows := []table.Row{}
	maxNameLen := 0
	for _, subtitle := range p.subtitles {
		if len(subtitle.Name) > maxNameLen {
			maxNameLen = len(subtitle.Name)
		}

		rows = append(rows, []string{
			fmt.Sprint(subtitle.SubtitleID),
			subtitle.Name,
			subtitle.Language,
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

func (p *OpenSubtitlesOrg) subtitleByID(id string) (providers.Subtitle, error) {
	for _, sub := range p.subtitles {
		if sub.SubtitleID == id {
			return sub, nil
		}
	}

	return nil, fmt.Errorf("subtitle with id %v not found in results", id)
}
