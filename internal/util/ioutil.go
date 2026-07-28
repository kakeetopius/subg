package util

import (
	"github.com/pterm/pterm"
)

func GetTermInput(prompt string, private bool) (string, error) {
	input := pterm.DefaultInteractiveTextInput.WithDelimiter(": ")
	if private {
		input.WithMask("*")
	}
	return input.Show(prompt)
}
