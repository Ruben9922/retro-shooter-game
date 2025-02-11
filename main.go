package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"os"
	"strings"
	"time"
)

var version = "dev"

type vector2d struct {
	x, y int
}

type model struct {
	windowSize vector2d
	view       view
	help       help.Model
}

type enemyTickMsg time.Time
type bulletTickMsg time.Time
type lifeLostTickMsg time.Time
type messageTickMsg time.Time

var emptyVector2d = vector2d{x: -1, y: -1}

var whiteColor = lipgloss.AdaptiveColor{
	Light: "8",
	Dark:  "7",
}
var blackColor = lipgloss.AdaptiveColor{
	Light: "15",
	Dark:  "0",
}
var accentColor = lipgloss.AdaptiveColor{
	Light: "12",
	Dark:  "4",
}
var secondaryTextStyle = help.New().Styles.ShortDesc

func initialModel() model {
	return model{
		view: newTitleView(),
		help: help.New(),
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func enemyTickCmd() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return enemyTickMsg(t)
	})
}

func bulletTickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return bulletTickMsg(t)
	})
}

func lifeLostTickCmd() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return lifeLostTickMsg(t)
	})
}

func messageTickCmd() tea.Cmd {
	return tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
		return messageTickMsg(t)
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowSize = vector2d{x: msg.Width, y: msg.Height}
	case tea.KeyMsg:
		return m.view.update(msg, m)
	case enemyTickMsg, bulletTickMsg, lifeLostTickMsg, messageTickMsg:
		return m.view.update(msg, m)
	}

	return m, nil
}

func (m model) View() string {
	titleBar := drawTitleBar(m, "Retro Shooter Game")
	mainView := lipgloss.PlaceHorizontal(
		m.windowSize.x,
		lipgloss.Center,
		lipgloss.NewStyle().Padding(1, 1).Render(m.view.draw(m)),
	)
	return lipgloss.NewStyle().Height(m.windowSize.y).Render(lipgloss.JoinVertical(lipgloss.Left, titleBar, mainView))
}

func drawTitleBar(m model, title string) string {
	horizontalPadding := 2
	titleBarStyle := lipgloss.NewStyle().Background(whiteColor).Foreground(blackColor).Bold(true).Padding(0, horizontalPadding)
	leftText := strings.Repeat(" ", lipgloss.Width(version))
	centerText := lipgloss.PlaceHorizontal(m.windowSize.x-(2*lipgloss.Width(version))-(horizontalPadding*2), lipgloss.Center, strings.ToUpper(title))
	return titleBarStyle.Render(lipgloss.JoinHorizontal(lipgloss.Top, leftText, centerText, version))
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
