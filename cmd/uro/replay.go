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

var recordDecorations = []string {
	" [rec] ",
	"       ",
}

var pauseDecorations = []string {
	" [pause] ",
	"         ",
}

func setupRecordReplay() error {
	err = nil
	if recordFile != "" {
		recorder, err = record.New()
	} else if replayFile != "" {
		player, err = record.Load(replayFile)
	}
	return err
}

func decorateFirstTab(title string) string {
	left := " "

	if recorder != nil {
		left = recordDecorations[t % 2]
	} else if player != nil {
		left = fmt.Sprintf(" [play %d%%] ", int(player.Progress()))
	}

	if paused && recorder == nil {
		left = pauseDecorations[t % 2]
	}

	return left + title
}
