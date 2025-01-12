package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type titleView struct{}

func newTitleView() titleView {
	return titleView{}
}

func (titleView) update(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "enter" {
			m.view = newGameView()
			return m, tea.Batch(enemyTickCmd(), bulletTickCmd())
		}
	}
	return m, nil
}

func (titleView) draw(model) string {
	titleStringPart1 := "  ____      _               ____  _                 _            \n |  _ \\ ___| |_ _ __ ___   / ___|| |__   ___   ___ | |_ ___ _ __ \n | |_) / _ \\ __| '__/ _ \\  \\___ \\| '_ \\ / _ \\ / _ \\| __/ _ \\ '__|\n |  _ <  __/ |_| | | (_) |  ___) | | | | (_) | (_) | ||  __/ |   \n |_| \\_\\___|\\__|_|  \\___/  |____/|_| |_|\\___/ \\___/ \\__\\___|_|   "
	titleStringPart2 := "   ____                      \n  / ___| __ _ _ __ ___   ___ \n | |  _ / _` | '_ ` _ \\ / _ \\\n | |_| | (_| | | | | | |  __/\n  \\____|\\__,_|_| |_| |_|\\___|"
	titleStringPart2WithVersion := lipgloss.JoinHorizontal(
		lipgloss.Bottom,
		lipgloss.NewStyle().MarginLeft(lipgloss.Width(version)+2).MarginRight(2).Render(titleStringPart2),
		secondaryTextStyle.Render(version),
	)
	titleString := lipgloss.JoinVertical(lipgloss.Center, titleStringPart1, titleStringPart2WithVersion)
	controlsString := "Press enter key to start..."
	viewString := lipgloss.JoinVertical(
		lipgloss.Center,
		lipgloss.NewStyle().PaddingBottom(2).Render(titleString),
		controlsString,
	)
	return lipgloss.NewStyle().Padding(1, 0).Render(viewString)
}
