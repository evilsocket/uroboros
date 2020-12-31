package views

import (
	"fmt"
	humanize "github.com/dustin/go-humanize"
	"github.com/evilsocket/uroboros/host"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"sort"
	"strings"
)

var processStates = map[string]string{
	"R": "Running",
	"S": "Sleeping in an interruptible wait",
	"D": "Waiting in uninterruptible disk sleep",
	"Z": "Zombie",
	"T": "Stopped (on a signal) or (before Linux 2.6.33) trace stopped",
	"t": "Tracing stop",
	"W": "Paging (< 2.6.0) or Waking",
	"X": "Dead",
	"x": "Dead",
	"K": "Wakekill",
	"P": "Parked",
}

// taken cursor https://pkg.go.dev/github.com/vosst/csi/proc/pid
const (
	PF_EXITING        = 0x00000004 // getting shut down
	PF_EXITPIDONE     = 0x00000008 // pi exit done on shut down
	PF_VCPU           = 0x00000010 // I'm a virtual CPU
	PF_WQ_WORKER      = 0x00000020 // I'm a workqueue worker
	PF_FORKNOEXEC     = 0x00000040 // forked but didn't exec
	PF_MCE_PROCESS    = 0x00000080 // process policy on mce errors
	PF_SUPERPRIV      = 0x00000100 // used super-user privileges
	PF_DUMPCORE       = 0x00000200 // dumped core
	PF_SIGNALED       = 0x00000400 // killed by a signal
	PF_MEMALLOC       = 0x00000800 // Allocating memory
	PF_NPROC_EXCEEDED = 0x00001000 // set_user noticed that RLIMIT_NPROC was exceeded
	PF_USED_MATH      = 0x00002000 // if unset the fpu must be initialized before use
	PF_USED_ASYNC     = 0x00004000 // used async_schedule*(), used by module init
	PF_NOFREEZE       = 0x00008000 // this thread should not be frozen
	PF_FROZEN         = 0x00010000 // frozen for system suspend
	PF_FSTRANS        = 0x00020000 // inside a filesystem transaction
	PF_KSWAPD         = 0x00040000 // I am kswapd
	PF_MEMALLOC_NOIO  = 0x00080000 // Allocating memory without IO involved
	PF_LESS_THROTTLE  = 0x00100000 // Throttle me less: I clean memory
	PF_KTHREAD        = 0x00200000 // I am a kernel thread
	PF_RANDOMIZE      = 0x00400000 // randomize virtual address space
	PF_SWAPWRITE      = 0x00800000 // Allowed to write to swap
	PF_NO_SETAFFINITY = 0x04000000 // Userland is not allowed to meddle with cpus_allowed
	PF_MCE_EARLY      = 0x08000000 // Early kill for mce process policy
	PF_MUTEX_TESTER   = 0x20000000 // Thread belongs to the rt mutex tester
	PF_FREEZER_SKIP   = 0x40000000 // Freezer should not count it as freezable
	PF_SUSPEND_TASK   = 0x80000000 // this thread called freeze_processes and should not be frozen
)

var processFlags = map[uint]string{
	PF_EXITING:        "EXITING",
	PF_EXITPIDONE:     "EXITPIDONE",
	PF_VCPU:           "VCPU",
	PF_WQ_WORKER:      "WQ_WORKER",
	PF_FORKNOEXEC:     "FORKNOEXEC",
	PF_MCE_PROCESS:    "MCE_PROCESS",
	PF_SUPERPRIV:      "SUPERPRIV",
	PF_DUMPCORE:       "DUMPCORE",
	PF_SIGNALED:       "SIGNALED",
	PF_MEMALLOC:       "MEMALLOC",
	PF_NPROC_EXCEEDED: "NPROC_EXCEEDED",
	PF_USED_MATH:      "USED_MATH",
	PF_USED_ASYNC:     "USED_ASYNC",
	PF_NOFREEZE:       "NOFREEZE",
	PF_FROZEN:         "FROZEN",
	PF_FSTRANS:        "FSTRANS",
	PF_KSWAPD:         "KSWAPD",
	PF_MEMALLOC_NOIO:  "MEMALLOC_NOIO",
	PF_LESS_THROTTLE:  "LESS_THROTTLE",
	PF_KTHREAD:        "KTHREAD",
	PF_RANDOMIZE:      "RANDOMIZE",
	PF_SWAPWRITE:      "SWAPWRITE",
	PF_NO_SETAFFINITY: "NO_SETAFFINITY",
	PF_MCE_EARLY:      "MCE_EARLY",
	PF_MUTEX_TESTER:   "MUTEX_TESTER",
	PF_FREEZER_SKIP:   "FREEZER_SKIP",
	PF_SUSPEND_TASK:   "SUSPEND_TASK",
}

func init() {
	registered["info"] = NewINFOView()
}

type INFOView struct {
	tree   *widgets.Tree
	table  *widgets.Table
	grid   *ui.Grid
	cursor int
	pid    int
	name   string
	parent string
}

func NewINFOView() *INFOView {
	v := INFOView{
		tree:  widgets.NewTree(),
		table: widgets.NewTable(),
		grid:  ui.NewGrid(),
	}

	v.tree.WrapText = false
	v.tree.SelectedRow = 1
	v.tree.SelectedRowStyle = ui.NewStyle(ui.ColorYellow, ui.ColorBlack, ui.ModifierBold)

	v.table.TextStyle = ui.NewStyle(ui.ColorWhite)
	v.table.RowSeparator = true
	v.table.FillRow = true
	v.table.Rows = [][]string{
		{"", ""},
	}
	v.table.ColumnResizer = v.setColumnSizes

	v.grid.Set(
		ui.NewRow(1.,
			ui.NewCol(1./5, v.tree),
			ui.NewCol(1.0-1./5, v.table),
		),
	)

	return &v
}

func (v *INFOView) Event(e ui.Event) {
	switch e.ID {
	case "j":
		v.tree.ScrollDown()
	case "k":
		v.tree.ScrollUp()
	case "<Enter>":
		if selected := v.tree.SelectedNode(); selected != nil {
			if n := selected.Value.(node); n.PID != 0 && n.PID != host.TargetPID {
				host.TargetPID = n.PID
			}
		}

	case "<Up>":
		if v.cursor > 0 {
			v.cursor--
		}
	case "<Down>":
		v.cursor++
	}
}

func (v *INFOView) AvailableFor(pid int) bool {
	return true
}

func (v *INFOView) Title() string {
	return fmt.Sprintf("(%d) %s", v.pid, v.name)
}

func (v *INFOView) setColumnSizes() {
	autosizeTable(v.table)
}

type node struct {
	PID  int
	Name string
}

func (n node) String() string {
	return n.Name
}

func (v *INFOView) updateTree(state *host.State) error {
	var main host.Task
	for _, task := range state.Process.Tasks {
		if task.ID == host.TargetPID {
			main = task
			break
		}
	}

	nodes := []*widgets.TreeNode{
		{
			Value: node{state.Process.Stat.PPID, v.parent},
			Nodes: []*widgets.TreeNode{
				{
					Value: node{main.ID, main.String()},
					Nodes: []*widgets.TreeNode{},
				},
			},
		},
	}

	for _, task := range state.Process.Tasks {
		if task.ID != host.TargetPID {
			nodes[0].Nodes[0].Nodes = append(nodes[0].Nodes[0].Nodes, &widgets.TreeNode{
				Value: node{task.ID, task.String()},
			})
		}
	}

	v.tree.SetNodes(nodes)
	v.tree.ExpandAll()

	return nil
}

func (v *INFOView) updateInfo(state *host.State) error {
	proc := state.Process
	stat := proc.Stat
	status := proc.Status

	flags := []string{}
	for mask, description := range processFlags {
		if stat.Flags&mask == mask {
			flags = append(flags, description)
		}
	}
	sort.Strings(flags)

	// UIDs of the process (Real, effective, saved set, and filesystem UIDs)
	types := []string{"real", "effective", "saved set", "filesystem"}
	users := []string{}
	for i, u := range proc.Users {
		users = append(users, fmt.Sprintf("%s: %s (%s)", types[i], u.Username, u.Uid))
	}

	// same with groups
	groups := []string{}
	for i, g := range proc.Groups {
		groups = append(groups, fmt.Sprintf("%s: %s (%s)", types[i], g.Name, g.Gid))
	}

	rows := [][]string{
		{" Start Time", fmt.Sprintf(" %s", proc.StartTime)},
		{" Running Time", fmt.Sprintf(" %s", state.ObservedAt.Sub(proc.StartTime))},
		{" Parent", fmt.Sprintf(" %s", v.parent)},
		{" PID", fmt.Sprintf(" %d", stat.PID)},
		{" Thread group ID", fmt.Sprintf(" %d", status.TGID)},
		{" User", fmt.Sprintf(" %s", strings.Join(users, " | "))},
		{" Group", fmt.Sprintf(" %s", strings.Join(groups, " | "))},
		{" Name", fmt.Sprintf(" %s", status.Name)},
		{" Executable", fmt.Sprintf(" %s", proc.Executable)},
		{" Command Line", fmt.Sprintf(" %s", strings.Join(proc.CmdLine, " "))},
		{" Root", fmt.Sprintf(" %s", proc.RootDir)},
		{" CWD", fmt.Sprintf(" %s", proc.Cwd)},
		{" Wait Channel", fmt.Sprintf(" %s", proc.WaitChan)},
		{" Session ID", fmt.Sprintf(" %d", stat.Session)},
		{" Priority", fmt.Sprintf(" %d", stat.Priority)},
		{" Nice", fmt.Sprintf(" %d", stat.Nice)},
		{" Threads", fmt.Sprintf(" %d", stat.NumThreads)},
		{" TTY", fmt.Sprintf(" %d", stat.TTY)},
		{" COMM", " " + stat.Comm},
		{" State", fmt.Sprintf(" %s (%s)", stat.State, processStates[stat.State])},
		{" Flags", fmt.Sprintf(" 0x%x %s", stat.Flags, strings.Join(flags, ", "))},
		{" Resident Mem", fmt.Sprintf(" %s", humanize.Bytes(uint64(stat.RSS*state.PageSize)))},
		{" Virtual Mem", fmt.Sprintf(" %s", humanize.Bytes(uint64(stat.VSize*uint(state.PageSize))))},
		{" Min Faults", fmt.Sprintf(" %d", stat.MinFlt)},
		{" Maj Faults", fmt.Sprintf(" %d", stat.MajFlt)},
	}

	if state.Offline {
		rows = append([][]string{{" Recording Time", fmt.Sprintf(" %s", state.ObservedAt)}}, rows...)
	}

	totRows := len(rows)
	hasScroll, scrollMsg, from, to := tableSetScroll(v.table, totRows, v.cursor)
	if hasScroll {
		v.cursor = from
		rows = rows[v.cursor:to]
		if len(rows) > 0 && len(rows[0]) > 0 {
			rows[0][0] += " " + scrollMsg
		}
	}

	v.table.Rows = rows

	v.pid = stat.PID
	v.name = status.Name

	return nil
}

func (v *INFOView) Update(state *host.State) error {
	if state.Process.Parent != nil && state.Process.ParentComm != "" {
		v.parent = fmt.Sprintf("(%d) %s", state.Process.Stat.PPID, state.Process.ParentComm)
	} else {
		v.parent = fmt.Sprintf("(%d)", state.Process.Stat.PPID)
	}

	if err := v.updateTree(state); err != nil {
		return err
	} else if err = v.updateInfo(state); err != nil {
		return err
	}

	return nil
}

func (v *INFOView) Drawable() ui.Drawable {
	return v.grid
}
