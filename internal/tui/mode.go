package tui

import tea "github.com/charmbracelet/bubbletea"

type modeModel struct {
	cursor int
}

func newModeModel() modeModel { return modeModel{} }

func (m modeModel) Init() tea.Cmd { return nil }

func (m modeModel) Update(msg tea.Msg) (modeModel, tea.Cmd) { return m, nil }

func (m modeModel) View() string { return "[Mode stub]" }
