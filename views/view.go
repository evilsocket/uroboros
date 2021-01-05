package views

import (
	"github.com/evilsocket/uroboros/host"
	ui "github.com/gizak/termui/v3"
)

var registered = map[string]View{}

type View interface {
	AvailableFor(pid int) bool
	Reset()
	Update(state *host.State) error
	Title() string
	Event(e ui.Event)
	Drawable() ui.Drawable
}

func ByName(name string) View {
	return registered[name]
}