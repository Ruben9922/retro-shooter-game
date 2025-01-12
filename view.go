package main

import tea "github.com/charmbracelet/bubbletea"

// todo: maybe just remove this and implement tea.Model instead?
type view interface {
	update(msg tea.Msg, m model) (tea.Model, tea.Cmd)
	draw(m model) string
}
