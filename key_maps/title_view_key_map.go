package key_maps

import "github.com/charmbracelet/bubbles/key"

type titleViewKeyMap struct {
	Start key.Binding
	Quit  key.Binding
}

var TitleViewKeys = titleViewKeyMap{
	Start: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("↵", "start"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q"),
		key.WithHelp("q", "quit"),
	),
}

func (k titleViewKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Start, k.Quit}
}

func (k titleViewKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Start, k.Quit}}
}
