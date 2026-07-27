package addic7ed

import (
	"context"
	"fmt"

	"github.com/kakeetopius/subg/internal/providers"
)

type Addic7ed struct{}

func (s A7Subtitle) ID() string {
	return fmt.Sprint(s.SubID)
}

func (s A7Subtitle) Fields() []string {
	return []string{fmt.Sprint(s.SubID), s.Name, s.Language, s.Version}
}

func NewProvider() *Addic7ed {
	return new(Addic7ed)
}

func (p *Addic7ed) Name() string {
	return "addic7ed.com"
}

func (p *Addic7ed) SearchSubtitle(ctx context.Context, opts providers.SearchOptions) ([]providers.Subtitle, error) {
	subs, err := searchSubtitle(opts)
	if err != nil {
		return nil, err
	}
	if len(subs) == 0 {
		return nil, providers.ErrNoResultsFound{Query: opts.Query, Episode: opts.Episode, Season: opts.Season}
	}
	return subs, nil
}

func (p *Addic7ed) SelectBest(subs []providers.Subtitle) providers.Subtitle {
	if len(subs) != 0 {
		return subs[0]
	}
	return A7Subtitle{}
}

func (p *Addic7ed) Download(ctx context.Context, sub providers.Subtitle) (providers.SubtitleFile, error) {
	subtitle := sub.(A7Subtitle)
	return downloadSubtitle(ctx, &subtitle)
}

func (p *Addic7ed) SubtitleHeaders() []string {
	return []string{"ID", "Name", "Lang", "Version"}
}
