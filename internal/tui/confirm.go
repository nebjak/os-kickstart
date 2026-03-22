package tui

import tea "github.com/charmbracelet/bubbletea"

type confirmModel struct {
	count int
	mode  mode
}

func newConfirmModel(count int, m mode) confirmModel {
	return confirmModel{
		count: count,
		mode:  m,
	}
}

func (m confirmModel) Init() tea.Cmd { return nil }

func (m confirmModel) Update(msg tea.Msg) (confirmModel, tea.Cmd) { return m, nil }

func (m confirmModel) View() string { return "[Confirm stub]" }
