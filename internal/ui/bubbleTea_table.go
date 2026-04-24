// Package ui is used to display various ui componenets to the terminal.
package ui

import (
	"errors"
	"fmt"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/kakeetopius/subg/internal/providers/addic7ed"
	"github.com/kakeetopius/subg/internal/providers/opensubtitles"
)

var ErrUserQuit = errors.New("user quit")

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type model struct {
	table         table.Model
	selectedRowID string
	userQuit      bool
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc":
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
		case "q", "ctrl+c":
			m.userQuit = true
			return m, tea.Quit
		case "enter":
			m.selectedRowID = m.table.SelectedRow()[0]
			return m, tea.Quit
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() tea.View {
	return tea.NewView(baseStyle.Render(m.table.View()) + "\n  " + m.table.HelpView() + "\n")
}

func DisplayOpenSubTable(subtitles []opensubtitles.OpenSubSubtitle) (*opensubtitles.OpenSubSubtitle, error) {
	if len(subtitles) < 1 {
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
	for _, subtitle := range subtitles {
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

	return openSubtitleObjByID(finalModel.selectedRowID, subtitles)
}

func DisplayAddic7edTable(subs *addic7ed.Addic7edSubtitle) (*addic7ed.SubtitleOption, error) {
	if len(subs.SubtitleOpts) == 0 {
		return nil, fmt.Errorf("no subtitles returned by addic7ed")
	}

	columns := []table.Column{
		{Title: "ID", Width: 5},
		{Title: "Name", Width: 70},
		{Title: "Lang", Width: 10},
		{Title: "Version", Width: 10},
	}

	rows := []table.Row{}
	for _, sub := range subs.SubtitleOpts {
		rows = append(rows, []string{
			fmt.Sprint(sub.ID),
			subs.Name,
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

	return addic7edSubtitleOptByID(finalModel.selectedRowID, subs)
}

func setUpTable(columns []table.Column, rows []table.Row, idenifierIndex int, tableWidth int) (tea.Model, error) {
	if len(columns) == 0 {
		return nil, fmt.Errorf("table columns empty")
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("table rows empty")
	}

	tableHeight := min(len(columns)+2, 10)
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(tableHeight),
		table.WithWidth(tableWidth),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	m := model{
		table:         t,
		selectedRowID: rows[0][idenifierIndex],
	}

	return m, nil
}

func openSubtitleObjByID(id string, subtitles []opensubtitles.OpenSubSubtitle) (*opensubtitles.OpenSubSubtitle, error) {
	for _, sub := range subtitles {
		if sub.SubtitleID == id {
			return &sub, nil
		}
	}

	return nil, fmt.Errorf("subtitle with id %v not found in results", id)
}

func addic7edSubtitleOptByID(id string, subtitle *addic7ed.Addic7edSubtitle) (*addic7ed.SubtitleOption, error) {
	for _, sub := range subtitle.SubtitleOpts {
		idStr := fmt.Sprint(id)
		if idStr == id {
			return &sub, nil
		}
	}

	return nil, fmt.Errorf("subtitle with id %v not found in results", id)
}
