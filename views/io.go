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
	t     int

	prevRChar uint64
	prevWChar uint64
	prevTime time.Time

	speed *widgets.Plot
	char  *widgets.Plot
	bytes *widgets.Plot
	grid  *ui.Grid
}

func NewIOView() *IOView {
	char := makeNLinesPlot(" total ", 2)
	speed := makeNLinesPlot(" speed ", 2)
	bytes := makeNLinesPlot(" storage ", 2)

	grid := ui.NewGrid()

	grid.Set(
		ui.NewRow(1.0/3,
			ui.NewCol(1.0, speed),
		),
		ui.NewRow(1.0/3,
			ui.NewCol(1.0, char),
		),
		ui.NewRow(1.0/3,
			ui.NewCol(1.0, bytes),
		),
	)

	return &IOView{
		char:  char,
		speed: speed,
		bytes: bytes,
		grid:  grid,
	}
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
	doReset := v.t >= pointsInTime(v.char)
	if doReset {
		v.t = 0
	}

	io := state.Process.IO

	perSec := 1. / state.ObservedAt.Sub(v.prevTime).Seconds()
	readSpeed := float64(io.RChar - v.prevRChar) * perSec
	writeSpeed := float64(io.WChar - v.prevWChar) * perSec

	updateNLinesPlot(v.speed, doReset,
		[]float64{readSpeed, writeSpeed},
		fmt.Sprintf(" speed (r:%s/s w:%s/s) ", humanize.Bytes(uint64(readSpeed)), humanize.Bytes(uint64(writeSpeed))))

	updateNLinesPlot(v.char, doReset,
		[]float64{float64(io.RChar), float64(io.WChar)},
		fmt.Sprintf(" total (r:%s w:%s) ", humanize.Bytes(io.RChar), humanize.Bytes(io.WChar)))

	updateNLinesPlot(v.bytes, doReset,
		[]float64{float64(io.ReadBytes), float64(io.WriteBytes)},
		fmt.Sprintf(" storage (r:%s w:%s) ", humanize.Bytes(io.ReadBytes), humanize.Bytes(io.WriteBytes)))

	v.t++
	v.prevRChar = io.RChar
	v.prevWChar = io.WChar
	v.prevTime = state.ObservedAt

	return nil
}

func (v *IOView) Drawable() ui.Drawable {
	if len(v.char.Data[0]) >= 2 {
		return v.grid
	}
	return empty
}
