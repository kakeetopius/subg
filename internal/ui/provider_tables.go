package ui

import (
	"fmt"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"github.com/kakeetopius/subg/internal/providers/addic7ed"
	"github.com/kakeetopius/subg/internal/providers/opensubtitles"
	"github.com/kakeetopius/subg/internal/providers/subdl"
)

func DisplayOpenSubTable(results opensubtitles.SearchResults) (*opensubtitles.Subtitle, error) {
	if len(results) < 1 {
		return nil, fmt.Errorf("subtitle results empty")
	}
	columns := []table.Column{
		{Title: "ID", Width: 8},
		{Title: "Name", Width: 72},
		{Title: "Lang", Width: 10},
		{Title: "Rating", Width: 10},
		{Title: "Votes", Width: 10},
	}

	rows := []table.Row{}
	for _, subtitle := range results {
		rows = append(rows, []string{
			subtitle.SubtitleID,
			subtitle.Release,
			subtitle.Language,
			fmt.Sprintf("%v", subtitle.Ratings),
			fmt.Sprintf("%v", subtitle.Votes),
		})
	}

	m, err := setUpTable(columns, rows, 0, 110)
	if err != nil {
		return nil, err
	}
	returnedModel, err := tea.NewProgram(m).Run()
	if err != nil {
		return nil, err
	}

	finalModel, ok := returnedModel.(model)
	if !ok {
		return nil, fmt.Errorf("could not get selected subtitle")
	}
	if finalModel.userQuit {
		return nil, ErrUserQuit
	}

	selectedSub, err := results.SubtitleByID(finalModel.selectedRowID)
	if err != nil {
		return nil, err
	}
	return selectedSub.(*opensubtitles.Subtitle), nil
}

func DisplayAddic7edTable(results *addic7ed.SearchResult) (*addic7ed.SubtitleOption, error) {
	if len(results.SubtitleOpts) == 0 {
		return nil, fmt.Errorf("no subtitles returned by addic7ed")
	}

	columns := []table.Column{
		{Title: "ID", Width: 5},
		{Title: "Name", Width: 70},
		{Title: "Lang", Width: 10},
		{Title: "Version", Width: 10},
	}

	rows := []table.Row{}
	for _, sub := range results.SubtitleOpts {
		rows = append(rows, []string{
			fmt.Sprint(sub.ID),
			results.Name,
			sub.Language,
			sub.Version,
		})
	}

	m, err := setUpTable(columns, rows, 0, 95)
	if err != nil {
		return nil, err
	}
	returnedModel, err := tea.NewProgram(m).Run()
	if err != nil {
		return nil, err
	}

	finalModel, ok := returnedModel.(model)
	if !ok {
		return nil, fmt.Errorf("could not get selected subtitle")
	}
	if finalModel.userQuit {
		return nil, ErrUserQuit
	}

	selectedSub, err := results.SubtitleByID(finalModel.selectedRowID)
	if err != nil {
		return nil, err
	}
	return selectedSub.(*addic7ed.SubtitleOption), nil
}

func DisplaySubDLTable(results *subdl.SearchResults) (*subdl.Subtitle, error) {
	if len(results.Subtitles) == 0 {
		return nil, fmt.Errorf("no subtitles returned by subdl")
	}

	columns := []table.Column{
		{Title: "ID", Width: 5},
		{Title: "Name", Width: 70},
		{Title: "Lang", Width: 10},
		{Title: "Author", Width: 15},
	}

	rows := []table.Row{}
	for _, sub := range results.Subtitles {
		rows = append(rows, []string{
			fmt.Sprint(sub.ID),
			sub.ReleaseName,
			sub.Lang,
			sub.Author,
		})
	}

	m, err := setUpTable(columns, rows, 0, 100)
	if err != nil {
		return nil, err
	}
	returnedModel, err := tea.NewProgram(m).Run()
	if err != nil {
		return nil, err
	}

	finalModel, ok := returnedModel.(model)
	if !ok {
		return nil, fmt.Errorf("could not get selected subtitle")
	}
	if finalModel.userQuit {
		return nil, ErrUserQuit
	}

	selectedSubtitle, err := results.SubtitleByID(finalModel.selectedRowID)
	if err != nil {
		return nil, err
	}
	return selectedSubtitle.(*subdl.Subtitle), nil
}
