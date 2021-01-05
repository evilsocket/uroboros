package views

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/evilsocket/uroboros/host"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/prometheus/procfs"
)

func init() {
	registered["maps"] = NewMAPSView()
}

type MAPSView struct {
	cursor int
	lastN int
	table  *widgets.Table
}

func NewMAPSView() *MAPSView {
	v := MAPSView{
		table: widgets.NewTable(),
	}

	v.Reset()

	return &v
}

func (v *MAPSView) Reset() {
	v.table.TextStyle = ui.NewStyle(ui.ColorWhite)
	v.table.RowSeparator = true
	v.table.FillRow = true
	v.table.Rows = [][]string{
		{" address range ", " perms ", " offset ", " dev ", " inode ", " path "},
	}
	v.table.RowStyles[0] = ui.NewStyle(ui.ColorWhite, ui.ColorBlack, ui.ModifierBold)
	v.table.ColumnResizer = v.setColumnSizes
}

func (v *MAPSView) Title() string {
	return fmt.Sprintf("maps %d", v.lastN)
}

func (v *MAPSView) AvailableFor(pid int) bool {
	return true
}

func (v *MAPSView) setColumnSizes() {
	autosizeTable(v.table)
}

func permsString(p *procfs.ProcMapPermissions) string {
	s := ""
	if p.Read {
		s += "r"
	} else {
		s += "-"
	}

	if p.Write {
		s += "w"
	} else {
		s += "-"
	}

	if p.Execute {
		s += "x"
	} else {
		s += "-"
	}

	if p.Private {
		s += "p"
	} else {
		s += "s"
	}

	return s
}

func (v *MAPSView) Event(e ui.Event) {
	switch e.ID {
	case "<Up>":
		if v.cursor > 0 {
			v.cursor--
		}
	case "<Down>":
		v.cursor++
	}
}

func (v *MAPSView) Update(state *host.State) error {
	rows := state.Process.Maps
	totRows := len(rows)
	v.lastN = totRows
	hasScroll, scrollMsg, from, to := tableSetScroll(v.table, totRows, v.cursor)

	if hasScroll {
		v.cursor = from
		rows = rows[v.cursor:to]
	}

	v.table.Rows = [][]string{
		{fmt.Sprintf(" address range %s", scrollMsg),
			" perms ",
			" offset ",
			" dev ",
			" inode ",
			" path "},
	}

	for _, entry := range rows {
		v.table.Rows = append(v.table.Rows, []string{
			fmt.Sprintf(" 0x%08x > 0x%08x (%s) ", entry.StartAddr, entry.EndAddr, humanize.Bytes(uint64(entry.EndAddr-entry.StartAddr))),
			fmt.Sprintf(" %s ", permsString(entry.Perms)),
			fmt.Sprintf(" %d ", entry.Offset),
			fmt.Sprintf(" 0x%x ", entry.Dev),
			fmt.Sprintf(" %d ", entry.Inode),
			fmt.Sprintf(" %s ", entry.Pathname)})
	}

	return nil
}

func (v *MAPSView) Drawable() ui.Drawable {
	return v.table
}
