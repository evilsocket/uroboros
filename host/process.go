package host

import (
	"github.com/prometheus/procfs"
	"io/ioutil"
	"os/user"
	"strconv"
	"strings"
	"time"
)

type Process struct {
	PID int

	Parent     *procfs.Proc
	ParentStat procfs.ProcStat
	ParentComm string

	Process     procfs.Proc
	Users       []*user.User
	Groups      []*user.Group
	Stat        procfs.ProcStat
	StartTime   time.Time
	Executable  string
	CmdLine     []string
	RootDir     string
	Cwd         string
	WaitChan    string
	MemoryLimit int
	Status      procfs.ProcStatus
	Maps        []*procfs.ProcMap
	FDs         procfs.ProcFDInfos
	FDInfos     map[string]FDInfo
	IO          procfs.ProcIO
	Tasks       []Task
}

func readIntFromFile(filepath string) (int, error) {
	s, err := ioutil.ReadFile(filepath)
	if err != nil {
		return -1, err
	}
	return strconv.Atoi(strings.TrimSpace(string(s)))
}

func parseProcess(pid int, procfs procfs.FS) (proc Process, err error) {
	var startTimeSecs float64

	proc = Process{
		PID:         pid,
		Users:       make([]*user.User, 0),
		Groups:      make([]*user.Group, 0),
		FDInfos:     make(map[string]FDInfo),
		MemoryLimit: -1, // unlimited by default
	}

	// gather the process specific info
	if proc.Process, err = procfs.Proc(pid); err != nil {
		return
	} else if proc.Stat, err = proc.Process.Stat(); err != nil {
		return
	}

	if startTimeSecs, err = proc.Stat.StartTime(); err != nil {
		return
	} else {
		proc.StartTime = time.Unix(int64(startTimeSecs), 0)
	}

	if proc.Executable, err = proc.Process.Executable(); err != nil {
		return
	} else if proc.CmdLine, err = proc.Process.CmdLine(); err != nil {
		return
	} else if proc.RootDir, err = proc.Process.RootDir(); err != nil {
		return
	} else if proc.Cwd, err = proc.Process.Cwd(); err != nil {
		return
	} else if proc.WaitChan, err = proc.Process.Wchan(); err != nil {
		return
	} else if proc.IO, err = proc.Process.IO(); err != nil {
		return
	}

	if cGroups, err := proc.Process.Cgroups(); err == nil {
		for _, cGroup := range cGroups {
			search := []string{
				"/sys/fs/cgroup" + cGroup.Path,
				"/sys/fs/cgroup/memory" + cGroup.Path,
			}

			for _, memCgroup := range search {
				if memoryLimit, err := readIntFromFile(memCgroup + "/memory.max"); err == nil {
					// v2 cgroup
					proc.MemoryLimit = memoryLimit
					break
				} else if memoryLimit, err = readIntFromFile(memCgroup + "/memory.limit_in_bytes"); err == nil {
					// v1 cgroup
					proc.MemoryLimit = memoryLimit
					break
				}
			}
		}
	}

	if proc.Status, err = proc.Process.NewStatus(); err != nil {
		return
	} else if proc.Maps, err = proc.Process.ProcMaps(); err != nil {
		return
	} else if proc.FDs, err = proc.Process.FileDescriptorsInfo(); err != nil {
		return
	} else if proc.Tasks, err = parseProcessTasks(pid); err != nil {
		return
	}

	for _, uid := range proc.Status.UIDs {
		var u *user.User
		if u, err = user.LookupId(uid); err != nil {
			proc.Users = append(proc.Users, &user.User{
				Uid:      uid,
				Username: uid,
			})
		} else {
			proc.Users = append(proc.Users, u)
		}
	}

	for _, gid := range proc.Status.GIDs {
		var g *user.Group
		if g, err = user.LookupGroupId(gid); err != nil {
			proc.Groups = append(proc.Groups, &user.Group{
				Gid:  gid,
				Name: gid,
			})
		} else {
			proc.Groups = append(proc.Groups, g)
		}
	}

	// and from its parent
	if parent, err := procfs.Proc(proc.Stat.PPID); err == nil {
		proc.Parent = &parent
		proc.ParentStat, _ = proc.Parent.Stat()
		proc.ParentComm, _ = proc.Parent.Comm()
	} else {
		proc.Parent = nil
	}

	return
}
