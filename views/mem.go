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
	viewWithPlots

	rss  *widgets.Plot
	virt *widgets.Plot
	swap *widgets.Plot
	grid *ui.Grid
	last float64
}

func NewMEMView() *MEMView {
	v := MEMView{
		rss:  widgets.NewPlot(),
		virt: widgets.NewPlot(),
		swap: widgets.NewPlot(),
		grid: ui.NewGrid(),
	}

	v.rss.Title = " mem usage "
	v.rss.AxesColor = ui.ColorWhite
	v.rss.Data = make([][]float64, 1)
	v.rss.Data[0] = []float64{}
	v.rss.MaxVal = 100.0

	v.virt.Title = " virtual memory "
	v.virt.AxesColor = ui.ColorWhite
	v.virt.LineColors = []ui.Color{ui.ColorGreen}
	v.virt.Data = make([][]float64, 1)

	v.swap.Title = " swap "
	v.swap.AxesColor = ui.ColorWhite
	v.swap.LineColors = []ui.Color{ui.ColorBlue}
	v.swap.Data = make([][]float64, 1)

	v.grid.Set(
		ui.NewRow(1.0/3,
			ui.NewCol(1.0, v.rss),
		),
		ui.NewRow(1.0/3,
			ui.NewCol(1.0, v.virt),
		),
		ui.NewRow(1.0/3,
			ui.NewCol(1.0, v.swap),
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
	v.trackUpdate(v.rss, v.virt, v.swap)

	memAvail := state.Memory.MemTotal * 1024
	used := uint64(state.Process.Stat.RSS * state.PageSize)
	usedPerc := float64(used) / float64(memAvail) * 100.0

	// check if we need to visualize the cgroup limit
	if state.Process.MemoryLimit > 0 && state.Process.MemoryLimit < memAvail {
		limitPerc := float64(state.Process.MemoryLimit) / float64(memAvail) * 100.0

		v.rss.Title = fmt.Sprintf(" resident memory - %s of %s (%.1f%%) - cgroup limit: %s ",
						humanize.Bytes(used),
						humanize.Bytes(memAvail),
						usedPerc,
						humanize.Bytes(state.Process.MemoryLimit))

		// if the limit just appeared we need to reallocate
		if len(v.rss.Data) == 1 {
			v.rss.Data = make([][]float64, 2)
			v.rss.LineColors = []ui.Color{ui.ColorRed, ui.ColorWhite}

			for numCurr := len(v.rss.Data[0]); len(v.rss.Data[1]) < numCurr; {
				v.rss.Data[1] = append(v.rss.Data[1], limitPerc)
			}
		}

		v.rss.Data[0] = append(v.rss.Data[0], usedPerc)
		v.rss.Data[1] = append(v.rss.Data[1], limitPerc)

	} else {
		v.rss.Title = fmt.Sprintf(" resident memory - %s of %s (%.1f%%) ",
						humanize.Bytes(used),
						humanize.Bytes(memAvail),
						usedPerc)

		if len(v.rss.Data) == 2 {
			v.rss.Data = make([][]float64, 1)
		}

		v.rss.Data[0] = append(v.rss.Data[0], usedPerc)
	}

	v.virt.Title = fmt.Sprintf(" virtual memory - %s ", humanize.Bytes(uint64(state.Process.Stat.VSize)))
	v.virt.Data[0] = append(v.virt.Data[0], float64(state.Process.Stat.VSize)/(1024*1024*1024))

	v.swap.Title = fmt.Sprintf(" swap - %s ", humanize.Bytes(uint64(state.Process.Status.VmSwap)))
	v.swap.Data[0] = append(v.swap.Data[0], float64(state.Process.Status.VmSwap)/(1024*1024*1024))

	v.last = usedPerc

	return nil
}

func (v *MEMView) Drawable() ui.Drawable {
	return v.grid
}
