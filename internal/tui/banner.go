package tui

import tea "github.com/charmbracelet/bubbletea"

type bannerModel struct{}

func newBannerModel() bannerModel { return bannerModel{} }

func (m bannerModel) Init() tea.Cmd { return nil }

func (m bannerModel) Update(msg tea.Msg) (bannerModel, tea.Cmd) {
	return m, nil
}

func (m bannerModel) View() string {
	return TitleStyle.Render("Kickstart") + "\n" + SubtitleStyle.Render("Loading...")
}
