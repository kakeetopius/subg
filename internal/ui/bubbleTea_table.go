// Package ui is used to display various ui componenets to the terminal.
package ui

import (
	"errors"
	"fmt"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
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
