package main

import (
	"fmt"
	"github.com/evilsocket/uroboros/host"
	"github.com/prometheus/procfs"
	"os"
	"sort"
	"strings"
)

func searchTarget() {
	if targetName != "" {
		if procs, err := procfs.AllProcs(); err != nil {
			panic(err)
		} else {
			matchPIDs := make([]int, 0)
			matches := make(map[int]procfs.Proc)
			for _, proc := range procs {
				if comm, _ := proc.Comm(); comm != "" && strings.Index(comm, targetName) != -1 {
					matches[proc.PID] = proc
					matchPIDs = append(matchPIDs, proc.PID)
				}
			}

			if num := len(matches); num == 0 {
				fmt.Printf("no matches for '%s'\n", targetName)
				os.Exit(1)
			} else if num > 1 {
				fmt.Printf("multiple matches for '%s':\n", targetName)

				sort.Ints(matchPIDs)

				for _, pid := range matchPIDs {
					proc := matches[pid]
					comm, _ := proc.Comm()
					cmdline, _ := proc.CmdLine()
					fmt.Printf("[%d] (%s) %s\n", pid, comm, strings.Join(cmdline, " "))
				}
				os.Exit(0)
			} else {
				for pid := range matches {
					host.TargetPID = pid
					return
				}
			}
		}
	}

	if host.TargetPID <= 0 {
		host.TargetPID = os.Getpid()
	}
}
