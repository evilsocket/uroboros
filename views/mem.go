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
	rss  *widgets.Plot
	virt *widgets.Plot
	grid *ui.Grid
	last float64
	t    int
}

func NewMEMView() *MEMView {
	v := MEMView{
		rss:  widgets.NewPlot(),
		virt: widgets.NewPlot(),
		grid: ui.NewGrid(),
	}

	v.rss.Title = " mem usage "
	v.rss.AxesColor = ui.ColorWhite
	v.rss.Data = make([][]float64, 1)
	v.rss.Data[0] = []float64{100.0}

	v.virt.Title = " virtual memory "
	v.virt.AxesColor = ui.ColorWhite
	v.virt.LineColors = []ui.Color{ui.ColorGreen}
	v.virt.Data = make([][]float64, 1)
	v.virt.Data[0] = []float64{0.0}

	v.grid.Set(
		ui.NewRow(1.0/2,
			ui.NewCol(1.0, v.rss),
		),
		ui.NewRow(1.0/2,
			ui.NewCol(1.0, v.virt),
		),
	)

	return &v
}

func (v *MEMView) AvailableFor(pid int) bool {
	return true
}

func (v *MEMView) Event(e ui.Event) {

}

func (v *MEMView) Title() string {
	return fmt.Sprintf("mem %.1f%%", v.last)
}

func (v *MEMView) Update(state *host.State) error {
	used := state.Process.Stat.RSS * state.PageSize
	usedPerc := float64(used) / float64(state.Memory.MemTotal*1024) * 100.0

	// TODO: unify this reset logic in a base class all views can use
	if v.t >= pointsInTime(v.rss) {
		v.t = 0
		v.rss.Data[0] = []float64{100.0}
		v.virt.Data[0] = []float64{0.0}
	}

	v.rss.Title = fmt.Sprintf(" resident memory - %s of %s (%.1f%%) ", humanize.Bytes(uint64(used)),
		humanize.Bytes(state.Memory.MemTotal*1024), usedPerc)
	v.rss.Data[0] = append(v.rss.Data[0], usedPerc)

	v.virt.Title = fmt.Sprintf(" virtual memory - %s ", humanize.Bytes(uint64(state.Process.Stat.VSize)))
	v.virt.Data[0] = append(v.virt.Data[0], float64(state.Process.Stat.VSize) / (1024*1024*1024))

	v.last = usedPerc
	v.t++

	return nil
}

func (v *MEMView) Drawable() ui.Drawable {
	return v.grid
}
