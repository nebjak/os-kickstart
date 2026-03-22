package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dpanic/os-kickstart/internal/modules"
)

type menuItem struct {
	module    modules.Module
	separator bool
	label     string
	Status    string
}

type menuModel struct {
	items     []menuItem
	allMods   []modules.Module
	cursor    int
	selected  map[int]bool
	height    int
	offset    int // scroll offset
	checksRan bool
}

func newMenuModel(mods []modules.Module) menuModel {
	var items []menuItem

	items = append(items, menuItem{separator: true, label: "Optimizations"})
	for _, m := range mods {
		if m.Category == "optimization" {
			items = append(items, menuItem{module: m})
		}
	}

	items = append(items, menuItem{separator: true, label: "Installations"})
	for _, sub := range modules.InstallSubsections() {
		hasItems := false
		for _, m := range mods {
			if m.Category == "installation" && m.Subsection == sub {
				hasItems = true
				break
			}
		}
		if !hasItems {
			continue
		}
		items = append(items, menuItem{separator: true, label: "  " + sub})
		for _, m := range mods {
			if m.Category == "installation" && m.Subsection == sub {
				items = append(items, menuItem{module: m})
			}
		}
	}

	cursor := 0
	for i, item := range items {
		if !item.separator {
			cursor = i
			break
		}
	}

	return menuModel{
		items:    items,
		allMods:  mods,
		cursor:   cursor,
		selected: make(map[int]bool),
		height:   30,
	}
}

func (m menuModel) Init() tea.Cmd {
	return runUpdateChecks(m.allMods)
}

func (m menuModel) Update(msg tea.Msg) (menuModel, tea.Cmd) {
	switch msg := msg.(type) {
	case updateCheckDoneMsg:
		m.checksRan = true
		for _, r := range msg.results {
			for i := range m.items {
				if !m.items[i].separator && m.items[i].module.ID == r.moduleID {
					m.items[i].Status = r.status
				}
			}
		}
		return m, nil

	case tea.WindowSizeMsg:
		// header=3 lines, footer=2 lines
		m.height = msg.Height - 5
		if m.height < 10 {
			m.height = 10
		}
		m.fixScroll()

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			m.cursor = m.prevSelectable(m.cursor)
			m.fixScroll()
		case "down", "j":
			m.cursor = m.nextSelectable(m.cursor)
			m.fixScroll()
		case " ":
			if !m.items[m.cursor].separator {
				if m.selected[m.cursor] {
					delete(m.selected, m.cursor)
				} else {
					m.selected[m.cursor] = true
				}
			}
		case "enter":
			selected := m.getSelected()
			if len(selected) == 0 {
				return m, nil
			}
			return m, tea.Batch(
				func() tea.Msg { return selectedModulesMsg{modules: selected} },
				func() tea.Msg { return switchScreenMsg{to: screenMode} },
			)
		case "q", "esc":
			return m, tea.Quit
		}
	}

	return m, nil
}

// fixScroll adjusts offset so cursor stays visible.
func (m *menuModel) fixScroll() {
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+m.height {
		m.offset = m.cursor - m.height + 1
	}
	if m.offset < 0 {
		m.offset = 0
	}
}

var (
	updateAvailableStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15"))
	latestStyle          = OKStyle
	installedStyle       = MutedStyle
	sectionStyle         = lipgloss.NewStyle().Bold(true).Foreground(ColorAccent)
	subsectionStyle      = lipgloss.NewStyle().Foreground(ColorAccent2)
)

func (m menuModel) View() string {
	var b strings.Builder

	// Fixed header
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(ColorAccent)
	b.WriteString(titleStyle.Render("  Kickstart"))
	if !m.checksRan {
		b.WriteString(MutedStyle.Render("  (checking for updates...)"))
	}
	b.WriteString("\n")
	b.WriteString(MutedStyle.Render("  ↑/↓ navigate • space select • enter confirm") + "\n\n")

	// Build all lines
	var lines []string
	for i, item := range m.items {
		lines = append(lines, m.renderItem(i, item))
	}

	// Apply scroll window
	end := m.offset + m.height
	if end > len(lines) {
		end = len(lines)
	}
	start := m.offset
	if start > len(lines) {
		start = len(lines)
	}

	// Scroll indicators
	if start > 0 {
		b.WriteString(MutedStyle.Render("  ▲ more above") + "\n")
	}

	for _, line := range lines[start:end] {
		b.WriteString(line + "\n")
	}

	if end < len(lines) {
		b.WriteString(MutedStyle.Render("  ▼ more below") + "\n")
	}

	// Footer
	count := len(m.selected)
	b.WriteString(MutedStyle.Render(fmt.Sprintf("\n  %d selected", count)))

	return b.String()
}

func (m menuModel) renderItem(i int, item menuItem) string {
	if item.separator {
		label := item.label
		if !strings.HasPrefix(label, "  ") {
			return sectionStyle.Render(fmt.Sprintf("  ── %s ──", label))
		}
		return subsectionStyle.Render(fmt.Sprintf("    %s", strings.TrimSpace(label)))
	}

	cursor := "  "
	if i == m.cursor {
		cursor = lipgloss.NewStyle().Foreground(ColorAccent).Render("▸ ")
	}

	checkbox := "[ ]"
	if m.selected[i] {
		checkbox = lipgloss.NewStyle().Foreground(ColorOK).Render("[✓]")
	}

	label := item.module.Label
	if i == m.cursor {
		label = lipgloss.NewStyle().Bold(true).Render(label)
	}

	line := fmt.Sprintf("%s%s %s", cursor, checkbox, label)

	if item.module.Description != "" {
		line += MutedStyle.Render(" — " + item.module.Description)
	}

	if item.Status != "" {
		switch {
		case strings.Contains(item.Status, "update available"):
			line += " " + updateAvailableStyle.Render(item.Status)
		case item.Status == "[latest]":
			line += " " + latestStyle.Render(item.Status)
		case item.Status == "[installed]":
			line += " " + installedStyle.Render(item.Status)
		default:
			line += " " + MutedStyle.Render(item.Status)
		}
	}

	return line
}

func (m menuModel) prevSelectable(from int) int {
	i := from - 1
	for i >= 0 {
		if !m.items[i].separator {
			return i
		}
		i--
	}
	i = len(m.items) - 1
	for i > from {
		if !m.items[i].separator {
			return i
		}
		i--
	}
	return from
}

func (m menuModel) nextSelectable(from int) int {
	i := from + 1
	for i < len(m.items) {
		if !m.items[i].separator {
			return i
		}
		i++
	}
	i = 0
	for i < from {
		if !m.items[i].separator {
			return i
		}
		i++
	}
	return from
}

func (m menuModel) getSelected() []modules.Module {
	var result []modules.Module
	for i, item := range m.items {
		if m.selected[i] && !item.separator {
			result = append(result, item.module)
		}
	}
	return result
}
