package views

import (
	"fmt"
	"github.com/evilsocket/uroboros/host"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"io/ioutil"
)

func init() {
	registered["stack"] = NewSTACKView()
}

type STACKView struct {
	tree   *widgets.Tree
	table  *widgets.Table
	grid   *ui.Grid
	cursor int
}

func NewSTACKView() *STACKView {
	v := STACKView{
		tree:  widgets.NewTree(),
		table: widgets.NewTable(),
		grid:  ui.NewGrid(),
	}

	v.Reset()

	return &v
}

func (v *STACKView) Reset() {
	v.tree.WrapText = false
	v.tree.SelectedRowStyle = ui.NewStyle(ui.ColorYellow, ui.ColorBlack, ui.ModifierBold)

	v.table.TextStyle = ui.NewStyle(ui.ColorWhite)
	v.table.RowSeparator = true
	v.table.FillRow = true
	v.table.Rows = [][]string{
		{"", "", "", ""},
	}
	v.table.ColumnResizer = v.setColumnSizes

	v.grid.Set(
		ui.NewRow(1.,
			ui.NewCol(1./5, v.tree),
			ui.NewCol(1.-1./5, v.table),
		),
	)
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
		v.tree.ScrollDown()
	case "k":
		v.tree.ScrollUp()
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
	var main host.Task
	for _, task := range state.Process.Tasks {
		if task.ID == host.TargetPID {
			main = task
			break
		}
	}

	nodes := []*widgets.TreeNode{
		{
			Value: node{main.ID, main.String()},
			Nodes: []*widgets.TreeNode{},
		},
	}

	for _, task := range state.Process.Tasks {
		if task.ID != host.TargetPID {
			nodes[0].Nodes = append(nodes[0].Nodes, &widgets.TreeNode{
				Value: node{task.ID, task.String()},
			})
		}
	}

	v.tree.SetNodes(nodes)
	v.tree.ExpandAll()

	var rows [][]string

	for i, task := range state.Process.Tasks {
		if i == v.tree.SelectedRow {
			for _, entry := range task.Stack {
				rows = append(rows, []string{
					fmt.Sprintf(" 0x%08x", entry.Address),
					fmt.Sprintf(" %s", entry.Function),
					fmt.Sprintf(" 0x%x", entry.Offset),
					fmt.Sprintf(" %d", entry.Size),
				})
			}
			break
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
