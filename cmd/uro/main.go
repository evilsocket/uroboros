package main

import (
	"flag"
	"fmt"
	"github.com/evilsocket/uroboros/host"
	ui "github.com/gizak/termui/v3"
	"os"
	"runtime/pprof"
	"time"
)

var err error

var cpuProfile = ""
var targetName = ""

var dataPeriod = 500
var viewPeriod = 500

func init() {
	flag.IntVar(&host.TargetPID, "pid", 0, "Process ID to monitor.")
	flag.StringVar(&targetName, "search", "", "Search target process by name.")
	flag.StringVar(&host.ProcFS, "procfs", host.ProcFS, "Root of the proc filesystem.")
	flag.StringVar(&tabIDS, "tabs", tabIDS, "Comma separated list of tab names to show.")

	flag.IntVar(&dataPeriod, "data-period", dataPeriod, "Data sample period in milliseconds.")
	flag.IntVar(&viewPeriod, "view-period", viewPeriod, "UI refresh period in milliseconds.")

	flag.StringVar(&recordFile, "record", recordFile, "If specified, record the session to this file.")
	flag.StringVar(&replayFile, "replay", replayFile, "If specified, replay the session in this file.")

	flag.StringVar(&cpuProfile, "cpu-profile", cpuProfile, "Used for debugging.")
}

func main() {
	flag.Parse()

	searchTarget()

	if cpuProfile != "" {
		f, err := os.Create(cpuProfile)
		if err != nil {
			panic(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if dataPeriod > viewPeriod {
		fmt.Println("The data period must be smaller or equal than the view period.")
		os.Exit(1)
	}

	if err = setupRecordReplay(); err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	if err = setupUI(host.TargetPID); err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	defer closeUI()

	uiEvents := ui.PollEvents()
	dataTicker := time.NewTicker(time.Millisecond * time.Duration(dataPeriod)).C
	viewTicker := time.NewTicker(time.Millisecond * time.Duration(viewPeriod)).C

	for {
		select {
		case <-dataTicker:
			sampleData()

		case <-viewTicker:
			updateTabs()
			renderUI()

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

			case "<Space>", "p":
				paused = !paused

			case "f":
				sampleData()
				updateTabs()
				renderUI()
			}

			// propagate to current view
			getActiveTab().Event(e)
		}
	}
}
