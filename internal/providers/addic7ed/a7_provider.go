package addic7ed

import (
	"fmt"

	"charm.land/bubbles/v2/table"
	"github.com/kakeetopius/subg/internal/providers"
	"github.com/kakeetopius/subg/internal/ui"
)

type Addic7ed struct {
	opts providers.Options

	subtitles []A7Subtitle
}

func (p *Addic7ed) Name() string {
	return "addic7ed"
}

func (p *Addic7ed) WithOptions(opts providers.Options) {
	p.opts = opts
}

func (p *Addic7ed) WithSpecificOptions(opts any) {
}

func (p *Addic7ed) SearchSubtitle() error {
	subs, err := searchSubtitle(p.opts)
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

func (p *Addic7ed) Download(subs []providers.Subtitle) error {
	for _, sub := range subs {
		subtitle := sub.(A7Subtitle)
		err := downloadSubtitle(&subtitle, p.opts)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Addic7ed) DownloadBest() error {
	return nil
}

func (p *Addic7ed) DisplaySelections() ([]providers.Subtitle, error) {
	if len(p.subtitles) == 0 {
		return nil, fmt.Errorf("no subtitles returned by addic7ed")
	}

	columns := []table.Column{
		{Title: "ID", Width: 5},
		{Title: "Name", Width: 70},
		{Title: "Lang", Width: 10},
		{Title: "Version", Width: 10},
	}

	rows := []table.Row{}
	for _, sub := range p.subtitles {
		rows = append(rows, []string{
			fmt.Sprint(sub.ID),
			p.opts.Query,
			sub.Language,
			sub.Version,
		})
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
