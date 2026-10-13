package opensubtitlesorg

import (
	"context"

	"github.com/kakeetopius/subg/internal/providers"
	"github.com/kakeetopius/subg/internal/sessions"
)

type OpenSubtitlesOrg struct {
	session sessions.Session
	OSOrgOptions
}

type OSOrgOptions struct {
	CacheDir string
}

type OSOrgSubtitle struct {
	Name       string
	SubtitleID string
	Release    string
	Votes      string
	Language   string
}

func (s OSOrgSubtitle) ID() string {
	return s.SubtitleID
}

func (s OSOrgSubtitle) Fields() []string {
	return []string{s.SubtitleID, s.Name, s.Language, s.Votes}
}

func NewProvider(opts OSOrgOptions) *OpenSubtitlesOrg {
	return &OpenSubtitlesOrg{
		OSOrgOptions: opts,
	}
}

func (p *OpenSubtitlesOrg) Name() string {
	return "opensubtitles.org"
}

func (p *OpenSubtitlesOrg) SearchSubtitle(ctx context.Context, opts providers.SearchOptions) ([]providers.Subtitle, error) {
	if opts.Season != 0 || opts.Episode != 0 {
		opts.IsSerie = true
	} else {
		opts.IsMovie = true
	}
	subs, err := p.search(opts)
	if err != nil {
		return nil, err
	}
	return subs, nil
}

func (p *OpenSubtitlesOrg) Download(ctx context.Context, sub providers.Subtitle) (providers.SubtitleFile, error) {
	subtitle := sub.(OSOrgSubtitle)
	return p.downloadSubtitle(ctx, &subtitle)
}

func (p *OpenSubtitlesOrg) SubtitleHeaders() []string {
	return []string{"ID", "Name", "Lang", "Votes"}
}
