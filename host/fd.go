package host

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"os"
	"strconv"
	"strings"
)

var stdIO = map[uint64]string{
	0: "0 (stdin)",
	1: "1 (stdout)",
	2: "2 (stderr)",
}

type FDType int

const (
	FDTypeFile FDType = iota
	FDTypeSocket
	FDTypeOther
)

type FDInfo struct {
	FD     uint64
	Name   string
	INode  uint64
	Type   FDType
	Target string
	Info   string
}

func resolveDescriptor(state *State, pid int, fdStr string) (fdInfo FDInfo, err error) {
	// get file descriptor number
	fd, err := strconv.ParseUint(fdStr, 10, 64)
	if err != nil {
		return
	}

	// resolve its symlink to the actual target
	path := fmt.Sprintf("%s/%d/fd/%d", ProcFS, pid, fd)
	target, e := os.Readlink(path)
	if e != nil {
		target = path
	}

	// check standards
	fdName := fdStr
	if std, found := stdIO[fd]; found {
		fdName = std
	}

	// TODO: add more parsers
	if idx := strings.Index(target, "socket:["); idx == 0 {
		fdInfo = FDInfo{
			Name: fdName,
			FD:   fd,
			Type: FDTypeSocket,
		}

		inodeStr := strings.TrimRight(target[8:], "]")
		fdInfo.INode, _ = strconv.ParseUint(inodeStr, 10, 64)

		// parse only if needed
		if state.NetworkINodes == nil {
			if state.NetworkINodes, err = parseNetworkInodes(); err != nil {
				return
			}
		}

		if entry, found := state.NetworkINodes[int(fdInfo.INode)]; found {
			fdInfo.Target = entry.String()
			fdInfo.Info = entry.InfoString()
		} else {
			fdInfo.Target = fmt.Sprintf("socket:[%d]", fdInfo.INode)
		}

	} else if idx := strings.Index(target, ":["); idx >= 0 {
		fdInfo = FDInfo{
			FD:     fd,
			Name:   fdName,
			Type:   FDTypeOther,
			Target: target,
		}
	} else {
		fdInfo = FDInfo{
			FD:     fd,
			Name:   fdName,
			Type:   FDTypeFile,
			Target: target,
		}
		if stat, err := os.Stat(target); err == nil {
			fdInfo.Info = stat.Mode().String()
			if sz := uint64(stat.Size()); sz > 0 {
				fdInfo.Info += " " + humanize.Bytes(sz)
			}
		}
	}

	return
}
