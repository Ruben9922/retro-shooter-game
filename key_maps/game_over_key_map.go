package key_maps

import "github.com/charmbracelet/bubbles/key"

type gameOverKeyMap struct {
	Restart key.Binding
	Quit    key.Binding
}

var GameOverKeys = gameOverKeyMap{
	Restart: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("↵", "restart"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q"),
		key.WithHelp("q", "quit"),
	),
}

func (k gameOverKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Restart, k.Quit}
}

func (k gameOverKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Restart, k.Quit}}
}
