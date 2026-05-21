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

var (
	ipStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))  // Cyan
	grayStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("245")) // Gray
	methodStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("214")) // Orange
	pathStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))  // Green
	status2xx   = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))  // Green
	status3xx   = lipgloss.NewStyle().Foreground(lipgloss.Color("226")) // Yellow
	status4xx   = lipgloss.NewStyle().Foreground(lipgloss.Color("196")) // Red
	status5xx   = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	userAgentStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240")) // Dark Gray
	refererStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("135")) // Purple
)

var logRegex = regexp.MustCompile(`^(\S+) (\S+) (\S+) \[(.*?)\] "(.*?)" (\d+) (\S+) "(.*?)" "(.*?)"`)

func highlightLine(line string) string {
	matches := logRegex.FindStringSubmatch(line)
	if len(matches) < 10 {
		return line
	}

	ip := ipStyle.Render(matches[1])
	ident := grayStyle.Render(matches[2])
	user := grayStyle.Render(matches[3])
	time := grayStyle.Render("[" + matches[4] + "]")
	
	request := matches[5]
	reqParts := strings.SplitN(request, " ", 3)
	if len(reqParts) == 3 {
		method := methodStyle.Render(reqParts[0])
		path := pathStyle.Render(reqParts[1])
		proto := reqParts[2]
		request = fmt.Sprintf("%s %s %s", method, path, proto)
	}
	request = "\"" + request + "\""

	statusStr := matches[6]
	status := statusStr
	switch {
	case strings.HasPrefix(statusStr, "2"):
		status = status2xx.Render(statusStr)
	case strings.HasPrefix(statusStr, "3"):
		status = status3xx.Render(statusStr)
	case strings.HasPrefix(statusStr, "4"):
		status = status4xx.Render(statusStr)
	case strings.HasPrefix(statusStr, "5"):
		status = status5xx.Render(statusStr)
	}

	size := grayStyle.Render(matches[7])
	referer := refererStyle.Render("\"" + matches[8] + "\"")
	ua := userAgentStyle.Render("\"" + matches[9] + "\"")

	return fmt.Sprintf("%s %s %s %s %s %s %s %s %s", ip, ident, user, time, request, status, size, referer, ua)
}

func (m model) getHighlightedContent() string {
	var sb strings.Builder
	for i, line := range m.filtered {
		sb.WriteString(highlightLine(line))
		if i < len(m.filtered)-1 {
			sb.WriteRune('\n')
		}
	}
	return sb.String()
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
	return header + "\n" + m.textInput.View() + "\n"
}
