package main

import (
	"fmt"
	"github.com/evilsocket/uroboros/record"
)

var recordFile = ""
var replayFile = ""
var recorder *record.Record
var player *record.Record
var paused = false

func decorateFirstTab(title string) string {
	// TODO: refactor this blinking shit
	// i really miss C's ternary operator :/
	left := " "
	if recorder != nil {
		if t%2 == 0 {
			left = " [rec] "
		} else {
			left = "       "
		}
	} else if player != nil {
		left = fmt.Sprintf(" [play %d%%] ", int(player.Progress()))
	}

	if paused && recorder == nil {
		left = " [pause] "
	}

	return left + title
}
