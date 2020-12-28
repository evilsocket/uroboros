package host

import (
	"github.com/prometheus/procfs"
	"os"
	"sync"
	"time"
)

var stateMutex = sync.Mutex{}
var state *State

type ProcessInfo struct {
	PID        int
	Parent     *procfs.Proc
	ParentStat *procfs.ProcStat
	Process    procfs.Proc
	Stat       procfs.ProcStat
	Status     procfs.ProcStatus
	Maps       []*procfs.ProcMap
	FDs        procfs.ProcFDInfos
	Stack      ProcessStack
}

type State struct {
	sync.Mutex
	// tcpAndUdpParser
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

func Observe(pid int) (*State, error) {
	var err error

	stateMutex.Lock()
	defer stateMutex.Unlock()

	if state == nil {
		state = &State{
			ObservedAt: time.Now(),
			PageSize:   os.Getpagesize(),
			Process: ProcessInfo{
				PID: pid,
			},
		}

		if state.procfs, err = procfs.NewDefaultFS(); err != nil {
			return nil, err
		}
	} else {
		state.ObservedAt = time.Now()
	}

	// gather host generic info first
	if state.Stat, err = state.procfs.Stat(); err != nil {
		return nil, err
	} else if state.Memory, err = state.procfs.Meminfo(); err != nil {
		return nil, err
	}

	// then gather the process specific info
	if state.Process.Process, err = state.procfs.Proc(pid); err != nil {
		return nil, err
	} else if state.Process.Stat, err = state.Process.Process.Stat(); err != nil {
		return nil, err
	} else if state.Process.Status, err = state.Process.Process.NewStatus(); err != nil {
		return nil, err
	} else if state.Process.Maps, err = state.Process.Process.ProcMaps(); err != nil {
		return nil, err
	} else if state.Process.FDs, err = state.Process.Process.FileDescriptorsInfo(); err != nil {
		return nil, err
	} else if state.Process.Stack, err = parseProcessStack(pid); err != nil {
		return nil, err
	}

	// and from its parent
	if parent, err := state.procfs.Proc(state.Process.Stat.PPID); err == nil {
		state.Process.Parent = &parent
		if parentStats, err := state.Process.Parent.Stat(); err == nil {
			state.Process.ParentStat = &parentStats
		} else {
			state.Process.ParentStat = nil
		}
	} else {
		state.Process.Parent = nil
	}

	if state.NetworkINodes, err = buildNetworkINodes(); err != nil {
		return nil, err
	}

	return state, err
}
