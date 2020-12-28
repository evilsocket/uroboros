package views

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/evilsocket/uroboros/host"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

func init() {
	registered["io"] = NewIOView()
}

type IOView struct {
	t     int
	char  *widgets.Plot
	sysc  *widgets.Plot
	bytes *widgets.Plot
	grid  *ui.Grid
}

func NewIOView() *IOView {
	char := makeNLinesPlot(" chars ", 2)
	sysc := makeNLinesPlot(" syscalls ", 2)
	bytes := makeNLinesPlot(" bytes ", 2)

	grid := ui.NewGrid()

	grid.Set(
		ui.NewRow(1.0/3,
			ui.NewCol(1.0, char),
		),
		ui.NewRow(1.0/3,
			ui.NewCol(1.0, sysc),
		),
		ui.NewRow(1.0/3,
			ui.NewCol(1.0, bytes),
		),
	)

	return &IOView{
		char:  char,
		sysc:  sysc,
		bytes: bytes,
		grid:  grid,
	}
}

func (v *IOView) Event(e ui.Event) {

}

func (v *IOView) Title() string {
	return "i/o"
}

func (v *IOView) Update(state *host.State) error {
	io, err := state.Process.Process.IO()
	if err != nil {
		return err
	}

	doReset := v.t >= pointsInTime
	if doReset {
		v.t = 0
	}

	updateNLinesPlot(v.char, doReset,
		[]float64{float64(io.RChar), float64(io.WChar)},
		fmt.Sprintf(" chars (r:%d w:%d) ", io.RChar, io.WChar))

	updateNLinesPlot(v.sysc, doReset,
		[]float64{float64(io.SyscR), float64(io.SyscW)},
		fmt.Sprintf(" syscalls (r:%d w:%d) ", io.SyscR, io.SyscW))

	updateNLinesPlot(v.bytes, doReset,
		[]float64{float64(io.ReadBytes), float64(io.WriteBytes)},
		fmt.Sprintf(" bytes (r:%s w:%s) ", humanize.Bytes(io.ReadBytes), humanize.Bytes(io.WriteBytes)))

	v.t++

	return nil
}

func (v *IOView) Render() ui.Drawable {
	return v.grid
}
