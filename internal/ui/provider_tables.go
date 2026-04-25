package ui

import (
	"fmt"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"github.com/kakeetopius/subg/internal/providers"
	"github.com/kakeetopius/subg/internal/providers/addic7ed"
	"github.com/kakeetopius/subg/internal/providers/opensubtitles"
	"github.com/kakeetopius/subg/internal/providers/subdl"
)

func DisplayOpenSubTable(results opensubtitles.SearchResults) (providers.Subtitle, error) {
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

	return displayTableAndGetSubtitle(results, rows, columns)
}

func DisplayAddic7edTable(results *addic7ed.SearchResult) (providers.Subtitle, error) {
	if len(results.Subtitles) == 0 {
		return nil, fmt.Errorf("no subtitles returned by addic7ed")
	}

	columns := []table.Column{
		{Title: "ID", Width: 5},
		{Title: "Name", Width: 70},
		{Title: "Lang", Width: 10},
		{Title: "Version", Width: 10},
	}

	rows := []table.Row{}
	for _, sub := range results.Subtitles {
		rows = append(rows, []string{
			fmt.Sprint(sub.ID),
			results.Name,
			sub.Language,
			sub.Version,
		})
	}

	return displayTableAndGetSubtitle(results, rows, columns)
}

func DisplaySubDLTable(results *subdl.SearchResults) (providers.Subtitle, error) {
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

	return displayTableAndGetSubtitle(results, rows, columns)
}

func displayTableAndGetSubtitle(results providers.SubtitleSearchResult, rows []table.Row, columns []table.Column) (providers.Subtitle, error) {
	m, err := setUpTable(columns, rows, 0)
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

	return results.SubtitleByID(finalModel.selectedRowID)
}
