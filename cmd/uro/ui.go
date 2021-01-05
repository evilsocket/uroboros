package main

import (
	"fmt"
	"github.com/evilsocket/islazy/str"
	"github.com/evilsocket/uroboros/host"
	"github.com/evilsocket/uroboros/views"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"os"
)

var tabIDS = "info, stack, cpu, mem, maps, io, fd"

var availTabIDS []string
var tabTitles []string
var tabViews []views.View

var grid *ui.Grid
var tabs *widgets.TabPane

var t = 0

func fatal(format string, a ...interface{}) {
	closeUI()
	fmt.Printf(format, a...)
	os.Exit(1)
}

func setupUI(pid int) error {
	if err := ui.Init(); err != nil {
		return err
	}

	termWidth, termHeight := ui.TerminalDimensions()

	grid = ui.NewGrid()
	grid.SetRect(0, 0, termWidth, termHeight)
	grid.Border = false

	for _, id := range str.Comma(tabIDS) {
		if view := views.ByName(id); view == nil {
			fatal("'%s' is not a valid tab name.\n", id)
		} else if !view.AvailableFor(pid) {
			fmt.Printf("tab '%s' is not available, try with sudo or another kernel.\n", id)
		} else {
			availTabIDS = append(availTabIDS, id)
			tabTitles = append(tabTitles, view.Title())
			tabViews = append(tabViews, view)
		}
	}

	if len(availTabIDS) == 0 {
		fatal("no tabs available.\n")
	}

	tabs = widgets.NewTabPane(tabTitles...)
	tabs.Border = false

	// most tabs need at least two data points to correctly render
	for i := 0; i < 2; i++ {
		sampleData()
		updateTabs()
	}

	// first render round
	renderUI()

	return nil
}

func closeUI() {
	ui.Close()

	if recorder != nil {
		fmt.Printf("saving session to %s ...\n", recordFile)
		if err = recorder.Save(recordFile); err != nil {
			panic(err)
		}
	} else if player != nil {
		var first, last host.State

		if err = player.First(&first); err != nil {
			fmt.Printf("%s over, error getting first frame: %v\n", replayFile, err)
		} else if err = player.Last(&last); err != nil {
			fmt.Printf("%s over, error getting last frame: %v\n", replayFile, err)
		} else {
			fmt.Printf("%s: replayed %d of %d frames for a total runtime of %v\n",
				replayFile,
				player.CurrentFrameIndex() + 1,
				player.TotalFrames(),
				last.ObservedAt.Sub(first.ObservedAt))
		}
	}
}

func getActiveTab() views.View {
	idx := 0
	if tabs.ActiveTabIndex > 0 && tabs.ActiveTabIndex < len(tabViews) {
		idx = tabs.ActiveTabIndex
	}
	return tabViews[idx]
}

func renderUI() {
	drawable := getActiveTab().Drawable()
	headRatio := 1. / 50

	if len(availTabIDS) == 1 && recorder == nil && player == nil {
		headRatio = 0.0
	}

	grid.Items = make([]*ui.GridItem, 0)
	grid.Set(
		ui.NewRow(headRatio,
			ui.NewCol(1.0, tabs),
		),
		ui.NewRow(1-headRatio,
			ui.NewCol(1.0, drawable),
		),
	)

	ui.Render(grid)
}

var prevPID = 0

func updateTabs() {
	if prevPID > 0 && prevPID != host.TargetPID {
		// the user selected a new process from the tree, views need reset
		for _, tab := range tabViews {
			tab.Reset()
		}
	}

	prevPID = host.TargetPID

	dataLock.Lock()
	defer dataLock.Unlock()

	for i, tab := range tabViews {
		// don't update the tab data if the user paused
		if !paused {
			if err = tab.Update(lastState); err != nil {
				fatal("error updating tab %s: %+v\n", availTabIDS[i], err)
			}
		}

		if i == 0 {
			tabs.TabNames[i] = decorateFirstTab(tab.Title())
		} else {
			tabs.TabNames[i] = tab.Title()
		}
	}

	t++
}
