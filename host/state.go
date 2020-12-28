package host

import (
	"github.com/prometheus/procfs"
	"sync"
	"time"
)

var stateMutex = sync.Mutex{}
var state *State

type ProcessInfo struct {
	PID int

	Parent     *procfs.Proc
	ParentStat *procfs.ProcStat

	Process procfs.Proc
	Stat    procfs.ProcStat
	Status  procfs.ProcStatus
	Maps    []*procfs.ProcMap
	FDs     procfs.ProcFDInfos
	Stack   ProcessStack
}

type State struct {
	sync.Mutex
	procfs procfs.FS
	// when this state has been parsed
	ObservedAt time.Time
	// network stuff
	NetworkINodes NetworkINodes
	// generic stats + cpu times
	PageSize int
	Stat     procfs.Stat
	Memory   procfs.Meminfo
	// process specific stats
	Process ProcessInfo
}
