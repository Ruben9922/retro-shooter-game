package main

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

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

func drawTitleBar(m model, title string) string {
	horizontalPadding := 2
	titleBarStyle := lipgloss.NewStyle().Background(whiteColor).Foreground(blackColor).Bold(true).Padding(0, horizontalPadding)
	leftText := strings.Repeat(" ", lipgloss.Width(version))
	centerText := lipgloss.PlaceHorizontal(m.windowSize.x-(2*lipgloss.Width(version))-(horizontalPadding*2), lipgloss.Center, strings.ToUpper(title))
	return titleBarStyle.Render(lipgloss.JoinHorizontal(lipgloss.Top, leftText, centerText, version))
}
