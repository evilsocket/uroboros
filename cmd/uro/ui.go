package main

import (
	"fmt"
	"github.com/evilsocket/uroboros/host"
	"github.com/evilsocket/uroboros/views"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

var tabIDs = []string{"info", "cpu", "stack", "mem", "maps", "io", "fd"}
var tabTitles = []string{}
var tabViews = []views.View{}
var numTabViews = len(tabIDs)

var currTabView views.View
var currTabDrawable ui.Drawable
var grid *ui.Grid
var tabs *widgets.TabPane

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

	headRatio := 0.04

	grid.Items = make([]*ui.GridItem, 0)
	grid.Set(
		ui.NewRow(headRatio,
			ui.NewCol(1.0, tabs),
		),
		ui.NewRow(1-headRatio,
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

