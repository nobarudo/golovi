package tui

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	ipStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))  // Cyan
	grayStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("245")) // Gray
	methodStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("214")) // Orange
	pathStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))  // Green
	status2xx      = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))  // Green
	status3xx      = lipgloss.NewStyle().Foreground(lipgloss.Color("226")) // Yellow
	status4xx      = lipgloss.NewStyle().Foreground(lipgloss.Color("196")) // Red
	status5xx      = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
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
