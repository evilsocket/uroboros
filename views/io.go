package views

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/evilsocket/uroboros/host"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"time"
)

func init() {
	registered["io"] = NewIOView()
}

type IOView struct {
	viewWithPlots

	prevRChar uint64
	prevWChar uint64
	prevTime  time.Time

	speed *widgets.Plot
	char  *widgets.Plot
	bytes *widgets.Plot
	grid  *ui.Grid
}

func NewIOView() *IOView {
	v := &IOView{
		grid: ui.NewGrid(),
	}

	v.Reset()

	return v
}

func (v *IOView) Reset() {
	v.char = makeNLinesPlot(" total ", 2, []ui.Color{ui.ColorGreen, ui.ColorRed})
	v.speed = makeNLinesPlot(" speed ", 2, []ui.Color{ui.ColorGreen, ui.ColorRed})
	v.bytes = makeNLinesPlot(" storage ", 2, []ui.Color{ui.ColorGreen, ui.ColorRed})

	v.grid.Set(
		ui.NewRow(1.0/3,
			ui.NewCol(1.0, v.speed),
		),
		ui.NewRow(1.0/3,
			ui.NewCol(1.0, v.char),
		),
		ui.NewRow(1.0/3,
			ui.NewCol(1.0, v.bytes),
		),
	)
}

func (v *IOView) Event(e ui.Event) {

}

func (v *IOView) AvailableFor(pid int) bool {
	return true
}

func (v *IOView) Title() string {
	return "i/o"
}

func (v *IOView) Update(state *host.State) error {
	v.trackUpdate(v.speed, v.char, v.bytes)

	io := state.Process.IO

	perSec := 1. / state.ObservedAt.Sub(v.prevTime).Seconds()
	readSpeed := float64(io.RChar-v.prevRChar) * perSec
	writeSpeed := float64(io.WChar-v.prevWChar) * perSec

	v.speed.Data[0] = append(v.speed.Data[0], readSpeed)
	v.speed.Data[1] = append(v.speed.Data[1], writeSpeed)
	v.speed.Title = fmt.Sprintf(" speed (r:%s/s w:%s/s) ", humanize.Bytes(uint64(readSpeed)), humanize.Bytes(uint64(writeSpeed)))

	v.char.Data[0] = append(v.char.Data[0], float64(io.RChar))
	v.char.Data[1] = append(v.char.Data[1], float64(io.WChar))
	v.char.Title = fmt.Sprintf(" total (r:%s w:%s) ", humanize.Bytes(io.RChar), humanize.Bytes(io.WChar))

	v.bytes.Data[0] = append(v.bytes.Data[0], float64(io.ReadBytes))
	v.bytes.Data[1] = append(v.bytes.Data[1], float64(io.WriteBytes))
	v.bytes.Title = fmt.Sprintf(" storage (r:%s w:%s) ", humanize.Bytes(io.ReadBytes), humanize.Bytes(io.WriteBytes))

	v.prevRChar = io.RChar
	v.prevWChar = io.WChar
	v.prevTime = state.ObservedAt

	return nil
}

func (v *IOView) Drawable() ui.Drawable {
	return v.grid
}
