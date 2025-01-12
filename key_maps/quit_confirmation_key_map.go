package key_maps

import "github.com/charmbracelet/bubbles/key"

type quitConfirmationKeyMap struct {
	Quit   key.Binding
	Cancel key.Binding
}

var QuitConfirmationKeys = quitConfirmationKeyMap{
	Quit: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("↵", "quit"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "cancel"),
	),
}

func (k quitConfirmationKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Quit, k.Cancel}
}

func (k quitConfirmationKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Quit, k.Cancel}}
}
