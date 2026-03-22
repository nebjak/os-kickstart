package tui

import "github.com/charmbracelet/lipgloss"

var (
	ColorAccent  = lipgloss.Color("212") // pink
	ColorAccent2 = lipgloss.Color("39")  // cyan
	ColorOK      = lipgloss.Color("78")  // green
	ColorWarn    = lipgloss.Color("208") // orange
	ColorError   = lipgloss.Color("196") // red
	ColorMuted   = lipgloss.Color("240") // gray

	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorAccent).
			MarginBottom(1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(ColorAccent2)

	OKStyle = lipgloss.NewStyle().
		Foreground(ColorOK)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorError)

	MutedStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	BorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorAccent).
			Padding(0, 1)
)
