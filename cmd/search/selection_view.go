// Package search is used to search and download subtitles.
package search

import (
	"fmt"
	"os"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/kakeetopius/subg/internal/providers/opensubtitles"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type model struct {
	table              table.Model
	selectedSubtitleID string
	userQuit           bool
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
			m.selectedSubtitleID = m.table.SelectedRow()[0]
			return m, tea.Quit
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() tea.View {
	return tea.NewView(baseStyle.Render(m.table.View()) + "\n  " + m.table.HelpView() + "\n")
}

func DisplaySubtitleTable(subtitles []opensubtitles.Subtitle) (*opensubtitles.Subtitle, error) {
	if len(subtitles) < 1 {
		return nil, fmt.Errorf("subtitle results empty")
	}
	columns := []table.Column{
		{Title: "ID", Width: 8},
		{Title: "Name", Width: 72},
		{Title: "Lang", Width: 10},
		{Title: "Rating", Width: 10},
	}

	rows := []table.Row{}
	for _, subtitle := range subtitles {
		rows = append(rows, []string{
			subtitle.SubtitleID,
			subtitle.Release,
			subtitle.Language,
			fmt.Sprintf("%v", subtitle.Ratings),
		})
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(7),
		table.WithWidth(100),
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
		table:              t,
		selectedSubtitleID: rows[0][0],
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
		os.Exit(1)
	}

	return subtitleObjByID(finalModel.selectedSubtitleID, subtitles)
}

func subtitleObjByID(id string, subtitles []opensubtitles.Subtitle) (*opensubtitles.Subtitle, error) {
	for _, sub := range subtitles {
		if sub.SubtitleID == id {
			return &sub, nil
		}
	}

	return nil, fmt.Errorf("subtitle with id %v not found in array", id)
}
