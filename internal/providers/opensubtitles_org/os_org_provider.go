package opensubtitlesorg

import (
	"context"
	"fmt"

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
	if len(subs) == 0 {
		if opts.IsSerie {
			return nil, fmt.Errorf("no results returned for %v Season %v Episode %v", opts.Query, opts.Season, opts.Episode)
		}
	}
	return subs, nil
}

func (p *OpenSubtitlesOrg) SelectBest(subs []providers.Subtitle) providers.Subtitle {
	if len(subs) != 0 {
		return subs[0]
	}
	return OSOrgSubtitle{}
}

func (p *OpenSubtitlesOrg) Download(ctx context.Context, sub providers.Subtitle) (providers.SubtitleFile, error) {
	subtitle := sub.(OSOrgSubtitle)
	return p.downloadSubtitle(ctx, &subtitle)
}

func (p *OpenSubtitlesOrg) SubtitleHeaders() []string {
	return []string{"ID", "Name", "Lang", "Votes"}
}
