package ui

import (
	"sync"
	"time"
)

var (
	ui       *TerminalUI
	progress *ProgressTracker
	initOnce sync.Once
)

func Init() {
	initOnce.Do(func() {
		progress = &ProgressTracker{
			startTime: time.Now(),
		}

		ui = &TerminalUI{
			eventQueue: make(chan UIEvent, 1000),
			done:       make(chan struct{}),
		}
		ui.StartLoop()
	})
}

func GetUI() *TerminalUI {
	return ui
}

func Stop() {
	ui.Stop()
}
