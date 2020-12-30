package views

import (
	"fmt"
	"github.com/evilsocket/uroboros/host"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"io/ioutil"
	"sort"
)

func init() {
	registered["stack"] = NewSTACKView()
}

type STACKView struct {
	list   *widgets.List
	table  *widgets.Table
	grid   *ui.Grid
	cursor int
}

func NewSTACKView() *STACKView {
	v := STACKView{
		list:  widgets.NewList(),
		table: widgets.NewTable(),
		grid:  ui.NewGrid(),
	}

	v.list.Title = "tasks (j/k)"
	v.list.WrapText = false
	v.list.SelectedRowStyle.Modifier = ui.ModifierBold
	v.list.SelectedRowStyle.Fg = ui.ColorYellow

	v.table.TextStyle = ui.NewStyle(ui.ColorWhite)
	v.table.RowSeparator = true
	v.table.FillRow = true
	v.table.Rows = [][]string{
		{"", "", "", ""},
	}
	v.table.ColumnResizer = v.setColumnSizes

	v.grid.Set(
		ui.NewRow(1.,
			ui.NewCol(1./15, v.list),
			ui.NewCol(1.-1./15, v.table),
		),
	)

	return &v
}

func (v *STACKView) Event(e ui.Event) {
	switch e.ID {
	case "<Up>":
		if v.cursor > 0 {
			v.cursor--
		}
	case "<Down>":
		v.cursor++

	case "j":
		v.list.ScrollDown()

	case "k":
		v.list.ScrollUp()
	}
}

func (v *STACKView) Title() string {
	return "stack"
}

func (v *STACKView) AvailableFor(pid int) bool {
	path := fmt.Sprintf("%s/%d/stack", host.ProcFS, pid)
	data, err := ioutil.ReadFile(path)
	return err == nil && len(data) > 0
}

func (v *STACKView) setColumnSizes() {
	autosizeTable(v.table)
}

func (v *STACKView) Update(state *host.State) error {
	sortedTaskIDS := make([]string, 0)
	for taskID := range state.Process.Stack {
		sortedTaskIDS = append(sortedTaskIDS, taskID)
	}
	sort.Strings(sortedTaskIDS)

	v.list.Rows = make([]string, len(sortedTaskIDS))
	for i, taskID := range sortedTaskIDS {
		v.list.Rows[i] = fmt.Sprintf(" %s ", taskID)
	}

	prevRows := v.table.Rows

	var rows [][]string

	for _, entry := range state.Process.Stack[sortedTaskIDS[v.list.SelectedRow]] {
		rows = append(rows, []string{
			fmt.Sprintf(" 0x%08x", entry.Address),
			fmt.Sprintf(" %s", entry.Function),
			fmt.Sprintf(" 0x%x", entry.Offset),
			fmt.Sprintf(" %d", entry.Size),
		})
	}

	for i, row := range prevRows {
		prevColVal := row[1]
		currColVal := v.table.Rows[i][1]
		// ignore changes on running time
		if i != 1 && prevColVal != currColVal {
			v.table.RowStyles[i] = ui.NewStyle(ui.ColorYellow, ui.ColorBlack, ui.ModifierBold)
		} else {
			v.table.RowStyles[i] = ui.NewStyle(ui.ColorWhite)
		}
	}

	totRows := len(rows)
	hasScroll, scrollMsg, from, to := tableSetScroll(v.table, totRows, v.cursor)
	if hasScroll {
		v.cursor = from
		rows = rows[v.cursor:to]
		if len(rows) > 0 && len(rows[0]) > 0 {
			rows[0][0] += scrollMsg
		}
	}

	v.table.Rows = [][]string{
		{fmt.Sprintf(" address %s", scrollMsg), " function", " offset", " size"},
	}

	v.table.Rows = append(v.table.Rows, rows...)

	return nil
}

func (v *STACKView) Drawable() ui.Drawable {
	return v.grid
}
