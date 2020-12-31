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
	plot    *widgets.Plot
	t       int
	last    float64
}

func NewCPUView() *CPUView {
	v := CPUView{
		plot: widgets.NewPlot(),
	}

	v.plot.Title = " cpu usage "
	v.plot.AxesColor = ui.ColorWhite
	v.plot.Data = make([][]float64, 1)
	v.plot.Data[0] = []float64{100.0}

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

	totalStime := float64(state.Process.Stat.STime - v.history.STime)
	totalUtime := float64(state.Process.Stat.UTime - v.history.UTime)
	total := totalStime + totalUtime

	total /= clockTicks
	seconds := state.ObservedAt.Sub(v.history.At).Seconds()
	cpu := (total / seconds) * 100.0

	v.history.Set(state)

	if v.t >= pointsInTime(v.plot) {
		v.t = 0
		v.plot.Data[0] = []float64{100.0}
	}

	v.last = cpu
	v.plot.Title = fmt.Sprintf(" cpu usage %.1f%% ", cpu)
	v.plot.Data[0] = append(v.plot.Data[0], cpu)
	v.t++

	return nil
}

func (v *CPUView) Drawable() ui.Drawable {
	if len(v.plot.Data[0]) >= 2 {
		return v.plot
	}
	return empty
}
