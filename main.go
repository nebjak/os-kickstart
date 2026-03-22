package main

import (
	"embed"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dpanic/os-kickstart/internal/tui"
)

//go:embed all:modules lib.sh
var assets embed.FS

var (
	version = "dev"
	commit  = "none"
)

func main() {
	m := tui.New(tui.Config{
		Assets:  assets,
		Version: version,
		Commit:  commit,
	})

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
