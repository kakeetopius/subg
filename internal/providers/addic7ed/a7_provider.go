package addic7ed

import (
	"context"
	"fmt"

	"charm.land/bubbles/v2/table"
	"github.com/kakeetopius/subg/internal/providers"
	"github.com/kakeetopius/subg/internal/ui"
)

type Addic7ed struct {
	subtitles []A7Subtitle
}

func NewProvider() *Addic7ed {
	return new(Addic7ed)
}

func (s A7Subtitle) IsSub() {}

func (p *Addic7ed) Name() string {
	return "addic7ed"
}

func (p *Addic7ed) SearchSubtitle(ctx context.Context, opts providers.Options) error {
	subs, err := searchSubtitle(opts)
	if err != nil {
		return err
	}
	if len(subs) == 0 {
		if opts.IsSerie {
			return fmt.Errorf("no results returned for %v Season %v Episode %v", opts.Query, opts.Season, opts.Episode)
		}
		return fmt.Errorf("no Results returned for %v", opts.Query)
	}
	p.subtitles = subs
	return nil
}

func (p *Addic7ed) Download(ctx context.Context, sub providers.Subtitle) (providers.SubtitleFile, error) {
	subtitle := sub.(A7Subtitle)
	return downloadSubtitle(ctx, &subtitle)
}

func (p *Addic7ed) DownloadBest(ctx context.Context) (providers.SubtitleFile, error) {
	return downloadSubtitle(ctx, &p.subtitles[0])
}

func (p *Addic7ed) PromptSelection() ([]providers.Subtitle, error) {
	if len(p.subtitles) == 0 {
		return nil, fmt.Errorf("no subtitles returned by addic7ed")
	}

	maxNameWidth := 70
	columns := []table.Column{
		{Title: "ID", Width: 5},
		{Title: "Name", Width: maxNameWidth},
		{Title: "Lang", Width: 10},
		{Title: "Version", Width: 10},
	}

	rows := []table.Row{}
	maxNameLen := 0
	for _, sub := range p.subtitles {
		if len(sub.Name) > maxNameLen {
			maxNameLen = len(sub.Name)
		}
		rows = append(rows, []string{
			fmt.Sprint(sub.ID),
			sub.Name,
			sub.Language,
			sub.Version,
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

func (p *Addic7ed) subtitleByID(id string) (providers.Subtitle, error) {
	for _, sub := range p.subtitles {
		idStr := fmt.Sprint(sub.ID)
		if idStr == id {
			return sub, nil
		}
	}

	return nil, fmt.Errorf("subtitle with id %v not found in results", id)
}
