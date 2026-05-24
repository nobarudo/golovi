package tui

import (
	"bufio"
	"os"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
)

type model struct {
	textInput  textinput.Model
	saveInput  textinput.Model
	viewport   viewport.Model
	allLogs    []string
	filtered   []string
	err        error
	ready      bool
	saving     bool
	lastSaved  string
}

func NewModel(filePath string) (model, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return model{}, err
	}
	defer file.Close()

	var logs []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		logs = append(logs, scanner.Text())
	}

	ti := textinput.New()
	ti.Placeholder = "Filter regex..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	si := textinput.New()
	si.Placeholder = "Enter filename (e.g. results.log)"
	si.CharLimit = 156
	si.Width = 40

	return model{
		textInput: ti,
		saveInput: si,
		allLogs:   logs,
		filtered:  logs,
	}, nil
}
