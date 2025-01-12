package key_maps

import "github.com/charmbracelet/bubbles/key"

type titleViewKeyMap struct {
	Start key.Binding
}

var TitleViewKeys = titleViewKeyMap{
	Start: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("↵", "start"),
	),
}

func (k titleViewKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Start}
}

func (k titleViewKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Start}}
}
