package key_maps

import "github.com/charmbracelet/bubbles/key"

type gameOverKeyMap struct {
	Restart key.Binding
}

var GameOverKeys = gameOverKeyMap{
	Restart: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("↵", "restart"),
	),
}

func (k gameOverKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Restart}
}

func (k gameOverKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Restart}}
}
