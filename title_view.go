package main

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"retro-shooter-game/key_maps"
)

type titleView struct{}

func newTitleView() titleView {
	return titleView{}
}

func (titleView) update(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, key_maps.TitleViewKeys.Start) {
			m.view = newGameView()
			return m, tea.Batch(enemyTickCmd(), bulletTickCmd())
		}
	}
	return m, nil
}

func (titleView) draw(m model) string {
	titleStringPart1 := "  ____      _               ____  _                 _            \n |  _ \\ ___| |_ _ __ ___   / ___|| |__   ___   ___ | |_ ___ _ __ \n | |_) / _ \\ __| '__/ _ \\  \\___ \\| '_ \\ / _ \\ / _ \\| __/ _ \\ '__|\n |  _ <  __/ |_| | | (_) |  ___) | | | | (_) | (_) | ||  __/ |   \n |_| \\_\\___|\\__|_|  \\___/  |____/|_| |_|\\___/ \\___/ \\__\\___|_|   "
	titleStringPart2 := "   ____                      \n  / ___| __ _ _ __ ___   ___ \n | |  _ / _` | '_ ` _ \\ / _ \\\n | |_| | (_| | | | | | |  __/\n  \\____|\\__,_|_| |_| |_|\\___|"
	titleStringPart2WithVersion := lipgloss.JoinHorizontal(
		lipgloss.Bottom,
		lipgloss.NewStyle().MarginLeft(lipgloss.Width(version)+2).MarginRight(2).Render(titleStringPart2),
		secondaryTextStyle.Render(version),
	)
	titleString := lipgloss.JoinVertical(lipgloss.Center, titleStringPart1, titleStringPart2WithVersion)
	pressToStartString := "Press enter key to start..."
	helpView := m.help.View(key_maps.TitleViewKeys)

	viewString := lipgloss.JoinVertical(
		lipgloss.Center,
		lipgloss.NewStyle().PaddingBottom(2).Render(titleString),
		lipgloss.NewStyle().PaddingBottom(1).Render(pressToStartString),
		helpView,
	)
	return lipgloss.NewStyle().Padding(1, 0).Render(viewString)
}
