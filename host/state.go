package host

import (
	"github.com/prometheus/procfs"
	"os"
	"sync"
	"time"
)

var CurrentState *State

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
	PID               int
	ParentProcess     *procfs.Proc
	ParentProcessStat *procfs.ProcStat
	Process           procfs.Proc
	ProcessStat       procfs.ProcStat
	ProcessStatus     procfs.ProcStatus
	ProcessMaps       []*procfs.ProcMap
	ProcessFDs        procfs.ProcFDInfos
	ProcessStack      ProcessStack
}

func Observe(pid int) (*State, error) {
	var err error

	// TODO: refactor this shit

	if CurrentState == nil {
		CurrentState = &State{
			PID: pid,
		}
		if CurrentState.procfs, err = procfs.NewDefaultFS(); err != nil {
			return nil, err
		}
	}

	CurrentState.Lock()
	defer CurrentState.Unlock()

	// gather host generic info first
	CurrentState.ObservedAt = time.Now()
	CurrentState.PageSize = os.Getpagesize()

	if CurrentState.Stat, err = CurrentState.procfs.Stat(); err != nil {
		return nil, err
	} else if CurrentState.Memory, err = CurrentState.procfs.Meminfo(); err != nil {
		return nil, err
	}

	// then gather the process specific info
	CurrentState.PID = pid
	if CurrentState.Process, err = CurrentState.procfs.Proc(pid); err != nil {
		return nil, err
	} else if CurrentState.ProcessStat, err = CurrentState.Process.Stat(); err != nil {
		return nil, err
	} else if CurrentState.ProcessStatus, err = CurrentState.Process.NewStatus(); err != nil {
		return nil, err
	} else if CurrentState.ProcessMaps, err = CurrentState.Process.ProcMaps(); err != nil {
		return nil, err
	} else if CurrentState.ProcessFDs, err = CurrentState.Process.FileDescriptorsInfo(); err != nil {
		return nil, err
	} else if CurrentState.ProcessStack, err = parseProcessStack(pid); err != nil {
		return nil, err
	}

	// and from its parent
	if parent, err := CurrentState.procfs.Proc(CurrentState.ProcessStat.PPID); err == nil {
		CurrentState.ParentProcess = &parent
		if parentStats, err := CurrentState.ParentProcess.Stat(); err == nil {
			CurrentState.ParentProcessStat = &parentStats
		} else {
			CurrentState.ParentProcessStat = nil
		}
	} else {
		CurrentState.ParentProcess = nil
	}

	if CurrentState.NetworkINodes, err = buildNetworkINodes(); err != nil {
		return nil, err
	}

	return CurrentState, err
}
