package key_maps

import "github.com/charmbracelet/bubbles/key"

type playingKeyMap struct {
	Left  key.Binding
	Right key.Binding
	Shoot key.Binding
	Pause key.Binding
}

var PlayingKeys = playingKeyMap{
	Left: key.NewBinding(
		key.WithKeys("left", "a"),
		key.WithHelp("←/a", "left"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "d"),
		key.WithHelp("→/d", "right"),
	),
	Shoot: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("␣", "shoot"),
	),
	Pause: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "pause"),
	),
}

func (k playingKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Left, k.Right, k.Shoot, k.Pause}
}

func (k playingKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Left, k.Right, k.Shoot, k.Pause}}
}
