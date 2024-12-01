package main

import tea "github.com/charmbracelet/bubbletea"

type view interface {
	update(msg tea.Msg, m model) (tea.Model, tea.Cmd)
	draw(m model) string
}
