package host

import (
	"github.com/prometheus/procfs"
	"os"
	"time"
)

var ProcFS = "/proc"
var TargetPID = 0

func Observe(pid int) (*State, error) {
	var err error

	stateMutex.Lock()
	defer stateMutex.Unlock()

	if state == nil {
		state = &State{
			Offline:    false,
			ObservedAt: time.Now(),
			PageSize:   os.Getpagesize(),
		}

		if state.procfs, err = procfs.NewFS(ProcFS); err != nil {
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
	if state.Process, err = parseProcess(pid, state.procfs); err != nil {
		return nil, err
	}

	// used to lookup fds/inodes
	if state.NetworkINodes, err = parseNetworkInodes(); err != nil {
		return nil, err
	}

	// lookup descriptors detailed info
	for _, fd := range state.Process.FDs {
		if info, err := resolveDescriptor(state, pid, fd.FD); err != nil {
			return nil, err
		} else {
			state.Process.FDInfos[fd.FD] = info
		}
	}

	return state, err
}
