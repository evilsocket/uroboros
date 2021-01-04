package views

import (
	"fmt"
	"github.com/evilsocket/uroboros/host"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"time"
)

func init() {
	registered["cpu"] = NewCPUView()
}

const clockTicks = 100.0

type cpuHistory struct {
	Inited bool
	At     time.Time
	STime  uint
	UTime  uint
}

func (h *cpuHistory) Set(state *host.State) {
	h.Inited = true
	h.At = state.ObservedAt
	h.STime = state.Process.Stat.STime
	h.UTime = state.Process.Stat.UTime
}

type CPUView struct {
	history cpuHistory
	tot     *widgets.Plot
	usr     *widgets.Plot
	sys     *widgets.Plot
	grid    *ui.Grid
	t       int
	last    float64
}

func NewCPUView() *CPUView {
	v := CPUView{
		tot:  widgets.NewPlot(),
		usr:  widgets.NewPlot(),
		sys:  widgets.NewPlot(),
		grid: ui.NewGrid(),
	}

	v.tot.Title = " total usage "
	v.tot.AxesColor = ui.ColorWhite
	v.tot.LineColors = []ui.Color{ui.ColorRed}
	v.tot.MaxVal = 100.0
	v.tot.Data = make([][]float64, 1)

	v.usr.Title = " user time "
	v.usr.AxesColor = ui.ColorWhite
	v.usr.LineColors = []ui.Color{ui.ColorYellow}
	v.usr.MaxVal = 100.0
	v.usr.Data = make([][]float64, 1)

	v.sys.Title = " kernel time "
	v.sys.AxesColor = ui.ColorWhite
	v.sys.LineColors = []ui.Color{ui.ColorWhite}
	v.sys.MaxVal = 100.0
	v.sys.Data = make([][]float64, 1)

	v.grid.Set(
		ui.NewRow(1.0/3,
			ui.NewCol(1.0, v.tot),
		),
		ui.NewRow(1.0/3,
			ui.NewCol(1.0, v.usr),
		),
		ui.NewRow(1.0/3,
			ui.NewCol(1.0, v.sys),
		),
	)


	return &v
}

func (v *CPUView) AvailableFor(pid int) bool {
	return true
}

func (v *CPUView) Event(e ui.Event) {

}

func (v *CPUView) Title() string {
	return fmt.Sprintf("cpu %.1f%%", v.last)
}

func (v *CPUView) Update(state *host.State) error {
	if !v.history.Inited {
		v.history.Set(state)
		return nil
	}

	sDelta := float64(state.Process.Stat.STime - v.history.STime)
	uDelta := float64(state.Process.Stat.UTime - v.history.UTime)
	total := sDelta + uDelta

	total /= clockTicks
	seconds := state.ObservedAt.Sub(v.history.At).Seconds()

	cpuUser := (uDelta / clockTicks / seconds) * 100.0
	cpuSys := (sDelta / clockTicks / seconds) * 100.0
	cpuTot := (total / seconds) * 100.0

	v.history.Set(state)

	if v.t >= pointsInTime(v.tot) {
		v.t = 0
		v.sys.Data[0] = []float64{}
		v.usr.Data[0] = []float64{}
		v.tot.Data[0] = []float64{}
	}

	v.last = cpuTot

	v.tot.Title = fmt.Sprintf(" total usage %.1f%% ", cpuTot)
	v.tot.Data[0] = append(v.tot.Data[0], cpuTot)
	v.usr.Title = fmt.Sprintf(" user time %.1f%% ", cpuUser)
	v.usr.Data[0] = append(v.usr.Data[0], cpuUser)
	v.sys.Title = fmt.Sprintf(" kernel time %.1f%% ", cpuSys)
	v.sys.Data[0] = append(v.sys.Data[0], cpuSys)

	v.t++

	return nil
}

func (v *CPUView) Drawable() ui.Drawable {
	if len(v.tot.Data[0]) >= 2 {
		return v.grid
	}
	return empty
}
