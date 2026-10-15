package opensubtitles

import (
	"context"
	"fmt"

	"github.com/kakeetopius/subg/internal/providers"
)

type OpenSubtitles struct {
	Options
}

type Options struct {
	APIKey   string
	CacheDir string
}

func (s OSSubtitle) ID() string {
	return s.SubtitleID
}

func (s OSSubtitle) Fields() []string {
	return []string{s.SubtitleID, s.Release, s.Language, fmt.Sprint(s.Ratings), fmt.Sprint(s.Votes)}
}

func NewProvider(opts Options) *OpenSubtitles {
	return &OpenSubtitles{
		Options: opts,
	}
}

func (p *OpenSubtitles) Name() string {
	return "opensubtitles.com"
}

func (p *OpenSubtitles) SearchSubtitle(ctx context.Context, opts providers.SearchOptions) ([]providers.Subtitle, error) {
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
		return nil, err
	}
	return subs, nil
}

func (p *OpenSubtitles) Download(ctx context.Context, sub providers.Subtitle) (providers.SubtitleFile, error) {
	subtitle := sub.(OSSubtitle)
	return p.downloadSubtitle(ctx, &subtitle)
}

func (p *OpenSubtitles) SubtitleHeaders() []string {
	return []string{"ID", "Name", "Lang", "Rating", "Votes"}
}
