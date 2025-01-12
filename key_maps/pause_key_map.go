package key_maps

import "github.com/charmbracelet/bubbles/key"

type pauseKeyMap struct {
	Resume key.Binding
	Quit   key.Binding
}

var PauseKeys = pauseKeyMap{
	Resume: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "resume"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q"),
		key.WithHelp("q", "quit"),
	),
}

func (k pauseKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Resume, k.Quit}
}

func (k pauseKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Resume, k.Quit}}
}
