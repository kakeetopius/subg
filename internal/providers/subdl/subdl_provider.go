// Package subdl is used to search for subtitles from subdl.com
package subdl

import (
	"context"
	"fmt"

	"github.com/kakeetopius/subg/internal/providers"
)

func (s SDSubtitle) ID() string {
	return fmt.Sprint(s.SubID)
}

func (s SDSubtitle) Fields() []string {
	return []string{s.SubID, s.Name, s.Author}
}

func NewProvider() *SubDL {
	return new(SubDL)
}

func (p *SubDL) Name() string {
	return "subdl.com"
}

func (p *SubDL) SearchSubtitle(ctx context.Context, searchOpts providers.SearchOptions) ([]providers.Subtitle, error) {
	// keyword "movie" or "tv" is what is required by the subdl API.
	featureType := "movie"
	if searchOpts.IsSerie || searchOpts.Episode != 0 || searchOpts.Season != 0 {
		// if a season or episode is given we assume it is a serie
		featureType = "tv"
	}
	searchOpts.Type = featureType

	subs, err := p.searchSubtitles(ctx, searchOpts)
	if err != nil {
		return nil, err
	}

	return subs, nil
}

func (p *SubDL) Download(ctx context.Context, sub providers.Subtitle) (providers.SubtitleFile, error) {
	subtitle := sub.(SDSubtitle)
	return p.downloadSubtitle(ctx, subtitle)
}

func (p *SubDL) SubtitleHeaders() []string {
	return []string{"ID", "Name", "Author"}
}
