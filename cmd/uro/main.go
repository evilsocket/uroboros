package main

import (
	"flag"
	"fmt"
	"github.com/evilsocket/uroboros/host"
	"github.com/evilsocket/uroboros/views"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"os"
	"time"
)

var targetPID = 0
var tabIDs = []string{"info", "cpu", "stack", "mem", "maps", "io", "fd"}
var tabTitles = []string{}
var tabViews = []views.View{}
var numTabViews = len(tabIDs)
var refreshPeriod = time.Duration(500 * time.Millisecond)
var currTabView views.View
var currTabDrawable ui.Drawable
var grid *ui.Grid
var tabs *widgets.TabPane

func init() {
	flag.IntVar(&targetPID, "pid", 0, "Process ID to monitor.")
}

func getActiveTab() views.View {
	idx := 0
	if tabs.ActiveTabIndex > 0 && tabs.ActiveTabIndex < numTabViews {
		idx = tabs.ActiveTabIndex
	}
	return tabViews[idx]
}

 func setupGrid() {
	grid.Lock()
	defer grid.Unlock()

	currTabView = getActiveTab()
	currTabDrawable = currTabView.Render()

	headRatio := 1.0 / 25

	grid.Items = make([]*ui.GridItem, 0)
	grid.Set(
		ui.NewRow(headRatio,
			ui.NewCol(1.0, tabs),
		),
		ui.NewRow(1 - headRatio,
			ui.NewCol(1.0, currTabDrawable),
		),
	)
}

func updateAllTabs() {
	if state, err := host.Observe(targetPID); err != nil {
		panic(err)
	} else {
		for i, tab := range tabViews {
			if err = tab.Update(state); err != nil {
				panic(fmt.Sprintf("error updating tab %s: %+v", tabIDs[i], err))
			}
			tabs.TabNames[i] = tab.Title()
		}
	}
}

func main() {
	flag.Parse()

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
	ticker := time.NewTicker(refreshPeriod).C

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
