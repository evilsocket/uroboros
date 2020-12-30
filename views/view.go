package views

import (
	"fmt"
	"github.com/evilsocket/uroboros/host"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

var empty = ui.NewCanvas()

var registered = map[string]View{}

type View interface {
	AvailableFor(pid int) bool
	Update(state *host.State) error
	Title() string
	Event(e ui.Event)
	Drawable() ui.Drawable
}

func ByName(name string) View {
	return registered[name]
}

func makeNLinesPlot(title string, n int) *widgets.Plot {
	plot := widgets.NewPlot()

	plot.Title = title
	plot.AxesColor = ui.ColorWhite
	plot.Data = make([][]float64, n)

	for i := 0; i < n; i++ {
		plot.Data[i] = []float64{}
	}

	return plot
}

func updateNLinesPlot(p *widgets.Plot, reset bool, data []float64, title string) {
	p.Lock()
	defer p.Unlock()

	p.Title = title

	if reset {
		n := len(data)
		for i := 0; i < n; i++ {
			// line plot requires at least two points
			p.Data[i] = []float64{p.Data[i][len(p.Data[i])-1], data[i]}
		}
	} else {
		for i, v := range data {
			p.Data[i] = append(p.Data[i], v)
		}
	}
}

func autosizeTable(table *widgets.Table) {
	if len(table.Rows) > 0 {
		maxs := make([]int, len(table.Rows[0]))
		for _, row := range table.Rows {
			for j, cell := range row {
				sz := len(cell)
				if sz > maxs[j] {
					maxs[j] = sz
				}
			}
		}

		totsz := 0
		for _, sz := range maxs {
			totsz += sz
		}

		availWidth := table.Dx()
		widths := []int{}
		for _, sz := range maxs {
			ratio := float64(sz) / float64(totsz)
			widths = append(widths, int(float64(availWidth)*ratio))
		}

		table.ColumnWidths = widths
	}
}

func tableVisibleRows(table *widgets.Table) int {
	// one row for data and one for divider
	return table.Inner.Dy() / 2
}

func tableSetScroll(table *widgets.Table, totRows int, cursor int) (bool, string, int, int) {
	table.Lock()
	defer table.Unlock()

	numVisibleRows := tableVisibleRows(table) - 1 // take header into account
	hasScroll := totRows > numVisibleRows && numVisibleRows > 0
	scroll := ""
	from := cursor
	to := 0

	if hasScroll {
		lastRowIdx := totRows - 1
		if from > lastRowIdx {
			from = lastRowIdx
		}

		to = from + numVisibleRows
		if to > lastRowIdx {
			to = lastRowIdx
		}

		scroll = fmt.Sprintf("(%c or %c to scroll) ", ui.UP_ARROW, ui.DOWN_ARROW)
	}

	return hasScroll, scroll, from, to
}

func pointsInTime(plot *widgets.Plot) int {
	dx := plot.Dx()
	if dx > 10 {
		return dx - 10
	}

	w, _ := ui.TerminalDimensions()
	return w
}