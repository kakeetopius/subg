package opensubtitles

import (
	"context"
	"fmt"

	"github.com/kakeetopius/subg/internal/providers"
)

type OpenSubtitles struct {
	OSOptions
}

type OSOptions struct {
	APIKey   string
	CacheDir string
}

func (s OSSubtitle) ID() string {
	return s.SubtitleID
}

func (s OSSubtitle) Fields() []string {
	return []string{s.SubtitleID, s.Release, s.Language, fmt.Sprint(s.Ratings), fmt.Sprint(s.Votes)}
}

func NewProvider(opts OSOptions) *OpenSubtitles {
	return &OpenSubtitles{
		OSOptions: opts,
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
	if len(subs) == 0 {
		if opts.IsSerie {
			return nil, fmt.Errorf("no results returned for %v Season %v Episode %v", opts.Query, opts.Season, opts.Episode)
		}
		return nil, fmt.Errorf("no Results returned for %v", opts.Query)
	}
	return subs, nil
}

func (p *OpenSubtitles) SelectBest(subs []providers.Subtitle) providers.Subtitle {
	if len(subs) != 0 {
		return subs[0]
	}
	return OSSubtitle{}
}

func (p *OpenSubtitles) Download(ctx context.Context, sub providers.Subtitle) (providers.SubtitleFile, error) {
	subtitle := sub.(OSSubtitle)
	return p.downloadSubtitle(ctx, &subtitle)
}

func (p *OpenSubtitles) SubtitleHeaders() []string {
	return []string{"ID", "Name", "Lang", "Rating", "Votes"}
}
