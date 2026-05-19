package tui

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	textInput  textinput.Model
	viewport   viewport.Model
	allLogs    []string
	filtered   []string
	err        error
	ready      bool
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

	return model{
		textInput: ti,
		allLogs:   logs,
		filtered:  logs,
	}, nil
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		headerHeight := 3 // title line + text input line + spacing
		footerHeight := 0
		verticalMarginHeight := headerHeight + footerHeight

		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			m.viewport.YPosition = headerHeight
			m.viewport.SetContent(strings.Join(m.filtered, "\n"))
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMarginHeight
		}
	}

	oldFilter := m.textInput.Value()
	m.textInput, tiCmd = m.textInput.Update(msg)
	
	if m.textInput.Value() != oldFilter {
		if m.textInput.Value() == "" {
			m.filtered = m.allLogs
		} else {
			re, err := regexp.Compile(m.textInput.Value())
			if err == nil {
				var filtered []string
				for _, line := range m.allLogs {
					if re.MatchString(line) {
						filtered = append(filtered, line)
					}
				}
				m.filtered = filtered
			}
		}
		m.viewport.SetContent(strings.Join(m.filtered, "\n"))
	}

	m.viewport, vpCmd = m.viewport.Update(msg)

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}
	return fmt.Sprintf(
		"%s\n%s",
		m.headerView(),
		m.viewport.View(),
	)
}

func (m model) headerView() string {
	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7D56F4")).
		Background(lipgloss.Color("#3C3C3C")).
		Padding(0, 1).
		Bold(true).
		Render(" GOLOVI - Nginx Log Viewer ")
	
	width := m.viewport.Width
	if width == 0 {
		width = 80
	}
	
	line := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#3C3C3C")).
		Render(strings.Repeat("─", width-lipgloss.Width(title)))
	
	header := lipgloss.JoinHorizontal(lipgloss.Bottom, title, line)
	return header + "\n" + m.textInput.View() + "\n"
}
