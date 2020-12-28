package views

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/evilsocket/uroboros/host"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"os"
	"sort"
	"strconv"
	"strings"
)

const (
	O_APPEND    = 0x400
	O_ASYNC     = 0x2000
	O_CLOEXEC   = 0x80000
	O_CREAT     = 0x40
	O_DIRECT    = 0x4000
	O_DIRECTORY = 0x10000
	O_DSYNC     = 0x1000
	O_EXCL      = 0x80
	O_LARGEFILE = 0x0
	O_NOATIME   = 0x40000
	O_NOCTTY    = 0x100
	O_NOFOLLOW  = 0x20000
	O_NONBLOCK  = 0x800
	O_PATH      = 0x200000
	O_SYNC      = 0x101000
	O_TMPFILE   = 0x410000
	O_TRUNC     = 0x200
)

var fdFlags = map[int]string{
	O_APPEND:    "APPEND",
	O_ASYNC:     "ASYNC",
	O_CLOEXEC:   "CLOEXEC",
	O_CREAT:     "CREAT",
	O_DIRECT:    "DIRECT",
	O_DIRECTORY: "DIRECTORY",
	O_DSYNC:     "DSYNC",
	O_EXCL:      "EXCL",
	O_LARGEFILE: "LARGEFILE",
	O_NOATIME:   "NOATIME",
	O_NOCTTY:    "NOCTTY",
	O_NOFOLLOW:  "NOFOLLOW",
	O_NONBLOCK:  "NONBLOCK",
	O_PATH:      "PATH",
	O_SYNC:      "SYNC",
	O_TMPFILE:   "TMPFILE",
	O_TRUNC:     "TRUNC",
}

var fdKnown = map[uintptr]string{
	0: "0 (stdin)",
	1: "1 (stdout)",
	2: "2 (stderr)",
}

func init() {
	registered["fd"] = NewFDView()
}

type FDView struct {
	t     int
	last   int
	cursor int
	plot   *widgets.Plot
	table  *widgets.Table
	grid   *ui.Grid
}

func NewFDView() *FDView {
	v := FDView{
		plot:  makeNLinesPlot(" fds ", 3),
		table: widgets.NewTable(),
		grid:  ui.NewGrid(),
	}

	v.table.TextStyle = ui.NewStyle(ui.ColorWhite)
	v.table.RowSeparator = true
	v.table.FillRow = true
	v.table.Rows = [][]string{
		{"", "", ""},
	}
	v.table.ColumnResizer = v.setColumnSizes

	v.grid.Set(
		ui.NewRow(1./4.,
			ui.NewCol(1.0, v.plot),
		),
		ui.NewRow(1 - 1/4.,
			ui.NewCol(1.0, v.table),
		),
	)

	return &v
}

func (v *FDView) Title() string {
	return fmt.Sprintf("fds %d", v.last)
}

func (v *FDView) setColumnSizes() {
	autosizeTable(v.table)
}

func resolveTargetFor(pid int, fd uintptr, state *host.State, numFiles *int, numSocks *int, numOther *int) (string, string, uint64) {
	perms := ""
	size := uint64(0)
	path := fmt.Sprintf("%s/%d/fd/%d", host.ProcFS, pid, fd)
	target, err := os.Readlink(path)
	if err != nil {
		target = path
	}

	if idx := strings.Index(target, "socket:["); idx == 0 {
		*numSocks++

		inodeStr := strings.TrimRight(target[8:], "]")
		inode, _ := strconv.ParseInt(inodeStr, 10, 64)
		if entry, found := state.NetworkINodes[int(inode)]; found {
			target = entry.String()
		} else {
			target = fmt.Sprintf("socket:[%d]", inode)
		}
	} else if idx := strings.Index(target, ":["); idx >= 0 {
		*numOther++
	} else {
		*numFiles++
		if info, err := os.Stat(target); err == nil {
			perms = info.Mode().String()
			size = uint64(info.Size())
		}
	}

	return target, perms, size
}

func (v *FDView) Event(e ui.Event) {
	switch e.ID {
	case "<Up>":
		if v.cursor > 0 {
			v.cursor--
		}
	case "<Down>":
		v.cursor++
	}
}

func (v *FDView) Update(state *host.State) error {
	var rows [][]string

	v.last = len(state.Process.FDs)

	numFiles := 0
	numSocks := 0
	numOther := 0

	for i, info := range state.Process.FDs {
		fdNum, _ := strconv.ParseUint(info.FD, 10, 32)
		fdName := info.FD
		if known, found := fdKnown[uintptr(fdNum)]; found {
			fdName = known
		}

		var flags []string
		for mask, desc := range fdFlags {
			flagsI, _ := strconv.ParseUint(state.Process.FDs[i].Flags, 16, 64)
			if flagsI&uint64(mask) == uint64(mask) {
				flags = append(flags, desc)
			}
		}
		sort.Strings(flags)

		target, perms, size := resolveTargetFor(state.Process.PID, uintptr(fdNum), state, &numFiles, &numSocks, &numOther)
		sizeStr := ""
		if size > 0 {
			sizeStr	= humanize.Bytes(size)
		}

		rows = append(rows, []string{
			fmt.Sprintf(" %s", fdName),
			fmt.Sprintf(" %s %s %s", target, perms, sizeStr),
			fmt.Sprintf(" 0x%s (%s)", state.Process.FDs[i].Flags, strings.Join(flags, ", ")),
		})
	}

	doReset := v.t >= pointsInTime(v.plot)
	if doReset {
		v.t = 0
	}

	updateNLinesPlot(v.plot, doReset,
		[]float64{float64(numFiles), float64(numSocks), float64(numOther)},
		fmt.Sprintf(" files:%d sockets:%d other:%d ", numFiles, numSocks, numOther))
	v.t++

	totRows := len(rows)
	hasScroll, scrollMsg, from, to := tableSetScroll(v.table, totRows, v.cursor)
	if hasScroll {
		v.cursor = from
		rows = rows[v.cursor:to]
	}

	v.table.Rows = [][]string{
		{fmt.Sprintf(" number %s", scrollMsg), " target", " flags"},
	}

	v.table.Rows = append(v.table.Rows, rows...)

	return nil
}

func (v *FDView) Render() ui.Drawable {
	return v.grid
}
