package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dpanic/os-kickstart/internal/modules"
)

type menuModel struct {
	items    []modules.Module
	cursor   int
	selected map[int]bool
}

func newMenuModel(mods []modules.Module) menuModel {
	return menuModel{
		items:    mods,
		selected: make(map[int]bool),
	}
}

func (m menuModel) Init() tea.Cmd { return nil }

func (m menuModel) Update(msg tea.Msg) (menuModel, tea.Cmd) { return m, nil }

func (m menuModel) View() string { return "[Menu stub]" }
