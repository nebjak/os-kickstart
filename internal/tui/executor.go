package tui

import tea "github.com/charmbracelet/bubbletea"

type executorModel struct{}

func newExecutorModel() executorModel { return executorModel{} }

func (m executorModel) Init() tea.Cmd { return nil }

func (m executorModel) Update(msg tea.Msg) (executorModel, tea.Cmd) { return m, nil }

func (m executorModel) View() string { return "[Executor stub]" }
