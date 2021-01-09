package views

import (
	"fmt"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

type viewWithPlots struct {
	t int
}

func pointsInTime(plot *widgets.Plot) int {
	dx := plot.Dx()
	if dx > 10 {
		return dx - 10
	}

	w, _ := ui.TerminalDimensions()
	return w
}

func (v *viewWithPlots) trackUpdate(plots ...*widgets.Plot) bool{
	first := plots[0]
	maxT := pointsInTime(first)

	if v.t >= maxT {
		v.t = 0
		for _, plot := range plots {
			for i, dataSoFar := range plot.Data {
				size := len(dataSoFar)
				last := 0.0
				if size >= 1 {
					last = dataSoFar[size - 1]
				}
				// we need to keep at least one
				plot.Data[i] = []float64{last}
			}
		}

		return true
	}

	v.t++

	return false
}

func makeNLinesPlot(title string, n int, lineColors []ui.Color) *widgets.Plot {
	plot := widgets.NewPlot()

	plot.Title = title
	plot.AxesColor = ui.ColorWhite
	plot.Data = make([][]float64, n)

	if lineColors != nil {
		plot.LineColors = lineColors
	}

	return plot
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
