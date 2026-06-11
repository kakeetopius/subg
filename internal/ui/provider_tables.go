package ui

import (
	"fmt"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
)

func DisplayTableAndGetSubtitleID(rows []table.Row, columns []table.Column) (string, error) {
	m, err := setUpTable(columns, rows, 0)
	if err != nil {
		return "", err
	}
	returnedModel, err := tea.NewProgram(m).Run()
	if err != nil {
		return "", err
	}

	finalModel, ok := returnedModel.(model)
	if !ok {
		return "", fmt.Errorf("could not get selected subtitle")
	}
	if finalModel.userQuit {
		return "", ErrUserQuit
	}
	if finalModel.nextProvider {
		return "", ErrNextProvider
	}

	return finalModel.selectedRowID, nil
}
