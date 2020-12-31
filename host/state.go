package host

import (
	"github.com/prometheus/procfs"
	"sync"
	"time"
)

var stateMutex = sync.Mutex{}
var state *State

type State struct {
	sync.Mutex
	Offline       bool // whether or not this state comes from a recording or it's live
	procfs        procfs.FS
	ObservedAt    time.Time     // when this state has been parsed
	NetworkINodes NetworkINodes // network stuff
	PageSize      int           // generic stats + cpu times
	Stat          procfs.Stat
	Memory        procfs.Meminfo
	Process       Process // process specific stats
}
