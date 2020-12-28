package views

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/evilsocket/uroboros/host"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

func init() {
	registered["mem"] = NewMEMView()
}

type MEMView struct {
	plot *widgets.Plot
	last float64
	t    int
}

func NewMEMView() *MEMView {
	v := MEMView{
		plot: widgets.NewPlot(),
	}

	v.plot.Title = " Memory Usage "
	v.plot.AxesColor = ui.ColorWhite
	v.plot.Data = make([][]float64, 1)
	v.plot.Data[0] = []float64{100.0}

	return &v
}

func (v *MEMView) Event(e ui.Event) {

}

func (v *MEMView) Title() string {
	return fmt.Sprintf("mem %.1f%%", v.last)
}

func (v *MEMView) Update(state *host.State) error {
	used := state.Process.Stat.RSS * state.PageSize
	usedPerc := float64(used) / float64(state.Memory.MemTotal * 1024) * 100.0

	if v.t >= pointsInTime(v.plot) {
		v.t = 0
		v.plot.Data[0] = []float64{100.0}
	}

	v.plot.Title = fmt.Sprintf(" Memory Usage %s of %s (%.1f%%) ", humanize.Bytes(uint64(used)),
		humanize.Bytes(state.Memory.MemTotal * 1024), usedPerc)
	v.plot.Data[0] = append(v.plot.Data[0], usedPerc)
	v.last = usedPerc
	v.t++

	return nil
}

func (v *MEMView) Render() ui.Drawable {
	return v.plot
}
