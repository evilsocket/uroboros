package main

import (
	"github.com/evilsocket/uroboros/host"
	"github.com/evilsocket/uroboros/record"
	"os"
	"sync"
)

var dataLock sync.Mutex
var lastState *host.State

func sampleData() {
	dataLock.Lock()
	defer dataLock.Unlock()

	// are we in replay mode?
	if player != nil {
		// and not in pause
		if !paused {
			// read the state from the next frame in the replay file
			var tmp host.State

			if err = player.Next(&tmp); err == record.EOF {
				closeUI()
				os.Exit(0)
			} else if err != nil {
				fatal("%v\n", err)
			} else {
				lastState = &tmp
				lastState.Offline = true
				host.TargetPID = tmp.Process.PID
			}
		}
	} else if lastState, err = host.Observe(host.TargetPID); err != nil {
		fatal("%v\n", err)
	}

	// save state if we're in record mode
	if recorder != nil {
		if err = recorder.Add(lastState); err != nil {
			fatal("%v\n", err)
		}
	}
}

