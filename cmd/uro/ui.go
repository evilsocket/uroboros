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

var tabIDS = "info, cpu, stack, mem, maps, io, fd"

var availTabIDS []string
var tabTitles []string
var tabViews []views.View

var grid *ui.Grid
var tabs *widgets.TabPane

func fatal(format string, a ...interface{}) {
	closeUI()
	fmt.Printf(format, a...)
	os.Exit(1)
}

func setupUI(pid int) error {
	if err := ui.Init(); err != nil {
		return err
	}

	grid = ui.NewGrid()
	termWidth, termHeight := ui.TerminalDimensions()
	grid.SetRect(0, 0, termWidth, termHeight)

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

	return nil
}

func closeUI() {
	ui.Close()
}

func getActiveTab() views.View {
	idx := 0
	if tabs.ActiveTabIndex > 0 && tabs.ActiveTabIndex < len(tabViews) {
		idx = tabs.ActiveTabIndex
	}
	return tabViews[idx]
}

func updateUI() {
	drawable := getActiveTab().Drawable()
	headRatio := 0.04

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

func updateTabs() {
	if state, err := host.Observe(targetPID); err != nil {
		fatal("%v\n", err)
	} else {
		for i, tab := range tabViews {
			if err = tab.Update(state); err != nil {
				fatal("error updating tab %s: %+v\n", tabIDS[i], err)
			}
			if i == 0 {
				tabs.TabNames[i] = " " + tab.Title()
			} else {
				tabs.TabNames[i] = tab.Title()
			}
		}
	}
}

