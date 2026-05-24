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

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		siCmd tea.Cmd
		vpCmd tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.saving {
			switch msg.Type {
			case tea.KeyEnter:
				filename := m.saveInput.Value()
				if filename == "" {
					filename = "output.log"
				}
				err := m.saveToFile(filename)
				if err != nil {
					m.lastSaved = "Error saving: " + err.Error()
				} else {
					m.lastSaved = "Saved to " + filename
				}
				m.saving = false
				m.saveInput.Blur()
				m.textInput.Focus()
				return m, nil
			case tea.KeyEsc, tea.KeyCtrlC:
				m.saving = false
				m.saveInput.Blur()
				m.textInput.Focus()
				return m, nil
			}
			m.saveInput, siCmd = m.saveInput.Update(msg)
			return m, siCmd
		}

		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyCtrlS:
			m.saving = true
			m.saveInput.Focus()
			m.saveInput.SetValue("output.log")
			m.textInput.Blur()
			return m, textinput.Blink
		}

	case tea.WindowSizeMsg:
		headerHeight := 4 // title line + text input line + status line + spacing
		footerHeight := 0
		verticalMarginHeight := headerHeight + footerHeight

		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			m.viewport.YPosition = headerHeight
			m.viewport.SetContent(m.getHighlightedContent())
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMarginHeight
		}
	}

	oldFilter := m.textInput.Value()
	m.textInput, tiCmd = m.textInput.Update(msg)

	if m.textInput.Value() != oldFilter {
		m.lastSaved = ""
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
		m.viewport.SetContent(m.getHighlightedContent())
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

	if m.saving {
		savePrompt := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFF00")).
			Bold(true).
			Render(" SAVE TO: ")
		return header + "\n" + m.textInput.View() + "\n" + savePrompt + m.saveInput.View() + "\n"
	}

	status := m.lastSaved
	if status == "" {
		status = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Ctrl+S: Save filtered results to output.log")
	} else {
		status = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render(status)
	}

	return header + "\n" + m.textInput.View() + "\n" + status + "\n"
}

func (m model) saveToFile(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	writer := bufio.NewWriter(f)
	for _, line := range m.filtered {
		_, err := writer.WriteString(line + "\n")
		if err != nil {
			return err
		}
	}
	return writer.Flush()
}
