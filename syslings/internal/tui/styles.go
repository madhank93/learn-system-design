package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Palette shared with kubelings/golings — the madhan.app house TUI style.
var (
	paneStyle   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("60")).Padding(0, 1)
	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("117"))
	headerStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("153"))
	moduleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("213"))
	cursorStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("231")).Background(lipgloss.Color("63"))
	textStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	dimStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	okStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("42"))
	warnStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214"))
	failStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("203"))
	linkStyle   = lipgloss.NewStyle().Underline(true).Foreground(lipgloss.Color("39"))

	keyStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("87"))
	sepStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	solvedStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("42"))
	startedStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214"))
	noneStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))

	fixStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("140")) // edit-the-config exercise
	scenarioStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("203")) // live-fault scenario
)

// keybar renders "key label · key label" with colored keys.
func keybar(pairs ...[2]string) string {
	var parts []string
	for _, p := range pairs {
		parts = append(parts, keyStyle.Render(p[0])+" "+textStyle.Render(p[1]))
	}
	return strings.Join(parts, sepStyle.Render(" · "))
}

func leftWidth(w int) int {
	lw := w * 4 / 10
	if lw < 30 {
		lw = 30
	}
	if lw > 48 {
		lw = 48
	}
	if lw > w-12 {
		lw = w - 12
	}
	return lw
}

func truncate(s string, w int) string {
	if w <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= w {
		return s
	}
	r := []rune(s)
	if len(r) <= w {
		return s
	}
	return string(r[:w-1]) + "…"
}

func padRight(s string, w int) string {
	d := w - lipgloss.Width(s)
	if d <= 0 {
		return s
	}
	return s + strings.Repeat(" ", d)
}
