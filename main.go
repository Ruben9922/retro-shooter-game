package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"os"
	"time"
)

type vector2d struct {
	x, y int
}

type model struct {
	windowSize vector2d
	view       view
}

type enemyTickMsg time.Time
type bulletTickMsg time.Time

var accentColor = lipgloss.AdaptiveColor{
	Light: "12",
	Dark:  "4",
}

func initialModel() model {
	return model{
		view: newGameView(),
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(bulletTickCmd(), enemyTickCmd())
}

func enemyTickCmd() tea.Cmd {
	return tea.Tick(900*time.Millisecond, func(t time.Time) tea.Msg {
		return enemyTickMsg(t)
	})
}

// todo: only use this if there are any bullets
func bulletTickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return bulletTickMsg(t)
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowSize = vector2d{x: msg.Width, y: msg.Height}
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		default:
			return m.view.update(msg, m)
		}
	case enemyTickMsg, bulletTickMsg:
		return m.view.update(msg, m)
	}

	return m, nil
}

func (m model) View() string {
	mainView := lipgloss.PlaceHorizontal(m.windowSize.x, lipgloss.Center,
		lipgloss.NewStyle().Padding(2, 4).Render(m.view.draw(m)))
	return lipgloss.NewStyle().Height(m.windowSize.y).Render(mainView)
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
