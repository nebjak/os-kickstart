package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dpanic/os-kickstart/internal/runner"
)

type summaryModel struct {
	results []runner.Result
}

func newSummaryModel(results []runner.Result) summaryModel {
	return summaryModel{results: results}
}

func (m summaryModel) Init() tea.Cmd { return nil }

func (m summaryModel) Update(msg tea.Msg) (summaryModel, tea.Cmd) { return m, nil }

func (m summaryModel) View() string { return "[Summary stub]" }
