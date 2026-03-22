package tui

import tea "github.com/charmbracelet/bubbletea"

type gitInfoModel struct{}

func newGitInfoModel(showWebhook bool) gitInfoModel { return gitInfoModel{} }

func (m gitInfoModel) Init() tea.Cmd { return nil }

func (m gitInfoModel) Update(msg tea.Msg) (gitInfoModel, tea.Cmd) { return m, nil }

func (m gitInfoModel) View() string { return "[GitInfo stub]" }
