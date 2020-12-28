package main

import (
	"flag"
	"fmt"
	"github.com/evilsocket/uroboros/views"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/prometheus/procfs"
	"os"
	"strings"
	"time"
)

var targetPID = 0
var targetName = ""
var refreshPeriod = 500

func init() {
	flag.IntVar(&targetPID, "pid", 0, "Process ID to monitor.")
	flag.StringVar(&targetName, "name", "", "Search target process by name.")
	flag.IntVar(&refreshPeriod, "period", refreshPeriod, "Data refresh period in milliseconds.")
}

func main() {
	flag.Parse()

	// TODO: handle errors with something better than a panic(err)

	if targetName != "" {
		if procs, err := procfs.AllProcs(); err !=  nil {
			panic(err)
		} else {
			matches := make(map[int]string)
			for _, proc := range procs {
				if comm, _ := proc.Comm(); comm != "" && strings.Index(comm, targetName) != -1 {
					matches[proc.PID] = comm
				}
			}

			if num := len(matches); num == 0 {
				fmt.Printf("no matches for '%s'\n", targetName)
				return
			} else if num > 1 {
				fmt.Printf("multiple matches for '%s':\n", targetName)
				for pid, comm := range matches {
					fmt.Printf("[%d] %s\n", pid, comm)
				}
				return
			} else {
				for pid := range matches {
					targetPID = pid
					break
				}
			}
		}
	}

	if targetPID <= 0 {
		targetPID = os.Getpid()
	}

	if err := ui.Init(); err != nil {
		panic(err)
	}
	defer ui.Close()

	grid = ui.NewGrid()
	termWidth, termHeight := ui.TerminalDimensions()
	grid.SetRect(0, 0, termWidth, termHeight)

	tabTitles = make([]string, numTabViews)
	for i, name := range tabIDs {
		v := views.ByName(name)
		tabViews = append(tabViews, v)
		tabTitles[i] = v.Title()
	}

	tabs = widgets.NewTabPane(tabTitles...)

	updateAllTabs()
	setupGrid()

	ui.Render(grid)

	uiEvents := ui.PollEvents()
	ticker := time.NewTicker(time.Millisecond * time.Duration(refreshPeriod)).C

	for {
		select {
		case <-ticker:
			updateAllTabs()

		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				return
			case "<Resize>":
				payload := e.Payload.(ui.Resize)
				grid.SetRect(0, 0, payload.Width, payload.Height)
			case "<Left>":
				tabs.FocusLeft()
			case "<Right>":
				tabs.FocusRight()
			}

			// propagate to current view
			getActiveTab().Event(e)
		}

		setupGrid()
		ui.Clear()
		ui.Render(grid)
	}
}
