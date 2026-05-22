package tui

import (
	"testing"
	"strings"
)

func TestHighlightLine(t *testing.T) {
	tests := []struct {
		name string
		line string
	}{
		{
			"Standard log",
			`85.208.96.211 - - [22/Mar/2026:00:00:11 +0900] "GET /robots.txt HTTP/1.1" 404 153 "-" "Mozilla/5.0" "-"`,
		},
		{
			"Success log",
			`127.0.0.1 - - [22/Mar/2026:00:00:11 +0900] "POST /api/data HTTP/1.1" 200 1234 "http://referer.com" "Mozilla/5.0" "-"`,
		},
		{
			"Server error log",
			`192.168.1.1 - admin [22/Mar/2026:00:00:11 +0900] "PUT /config HTTP/1.1" 500 0 "-" "-" "-"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			highlighted := highlightLine(tt.line)
			// We don't strictly check for ANSI codes because lipgloss might suppress them in non-TTY environments.
			// Instead, we check that the function doesn't return an empty string and preserves the original data.
			if highlighted == "" {
				t.Errorf("highlightLine returned an empty string")
			}
			
			// Verify that the key components are still present
			if !strings.Contains(highlighted, "2026") {
				t.Errorf("Highlighted line lost the date")
			}
		})
	}
}
