package subdl

import (
	"context"
	"fmt"

	"charm.land/bubbles/v2/table"
	"github.com/kakeetopius/subg/internal/providers"
	"github.com/kakeetopius/subg/internal/ui"
)

type SubDL struct {
	subtitles []SDSubtitle
	SubDLOpts
}

type SubDLOpts struct {
	APIKey string
}

func (s SDSubtitle) IsSub() {}

func NewProvider(opts SubDLOpts) *SubDL {
	return &SubDL{
		SubDLOpts: opts,
	}
}

func (p *SubDL) Name() string {
	return "subdl.com"
}

func (p *SubDL) SearchSubtitle(ctx context.Context, searchOpts providers.Options) error {
	// keyword "movie" or "tv" is what is required by the subdl API.
	featureType := "movie"
	if searchOpts.IsSerie || searchOpts.Episode != 0 || searchOpts.Season != 0 {
		// if a season or episode is given we assume it is a serie
		featureType = "tv"
	}
	searchOpts.Type = featureType

	subs, err := p.searchSubtitles(searchOpts)
	if err != nil {
		return err
	}

	if len(subs) == 0 {
		if searchOpts.IsSerie {
			return fmt.Errorf("no results returned for %v Season %v Episode %v", searchOpts.Query, searchOpts.Season, searchOpts.Episode)
		}
		return fmt.Errorf("no Results returned for %v", searchOpts.Query)
	}
	p.subtitles = subs
	return nil
}

func (p *SubDL) Download(ctx context.Context, sub providers.Subtitle) (providers.SubtitleFile, error) {
	subtitle := sub.(SDSubtitle)
	return downloadSubtitle(ctx, &subtitle)
}

func (p *SubDL) DownloadBest(ctx context.Context) (providers.SubtitleFile, error) {
	return downloadSubtitle(ctx, &p.subtitles[0])
}

func (p *SubDL) PromptSelection() ([]providers.Subtitle, error) {
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
