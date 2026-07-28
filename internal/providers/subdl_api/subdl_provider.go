package subdlapi

import (
	"context"
	"fmt"

	"github.com/kakeetopius/subg/internal/providers"
)

type SubDL struct {
	SubDLOpts
}

type SubDLOpts struct {
	APIKey string
}

func (s SDSubtitle) ID() string {
	return fmt.Sprint(s.SubID)
}

func (s SDSubtitle) Fields() []string {
	return []string{fmt.Sprint(s.SubID), s.ReleaseName, s.Lang, s.Author}
}

func NewProvider(opts SubDLOpts) *SubDL {
	return &SubDL{
		SubDLOpts: opts,
	}
}

func (p *SubDL) Name() string {
	return "subdl.com via the api"
}

func (p *SubDL) SearchSubtitle(ctx context.Context, searchOpts providers.SearchOptions) ([]providers.Subtitle, error) {
	// keyword "movie" or "tv" is what is required by the subdl API.
	featureType := "movie"
	if searchOpts.IsSerie || searchOpts.Episode != 0 || searchOpts.Season != 0 {
		// if a season or episode is given we assume it is a serie
		featureType = "tv"
	}
	searchOpts.Type = featureType

	subs, err := p.searchSubtitles(searchOpts)
	if err != nil {
		return nil, err
	}

	if len(subs) == 0 {
		return nil, providers.ErrNoResultsFound{Query: searchOpts.Query, Episode: searchOpts.Episode, Season: searchOpts.Season}
	}
	return subs, nil
}

func (p *SubDL) SelectBest(subs []providers.Subtitle) providers.Subtitle {
	if len(subs) != 0 {
		return subs[0]
	}
	return SDSubtitle{}
}

func (p *SubDL) Download(ctx context.Context, sub providers.Subtitle) (providers.SubtitleFile, error) {
	subtitle := sub.(SDSubtitle)
	return downloadSubtitle(ctx, &subtitle)
}

func (p *SubDL) SubtitleHeaders() []string {
	return []string{"ID", "Name", "Lang", "Author"}
}
