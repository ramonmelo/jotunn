package ui

import (
	"fmt"
	"os"
	"sync"
	"time"
)

type TerminalUI struct {
	mu           sync.Mutex
	eventQueue   chan UIEvent
	progressSize int
}

type UIEvent struct {
	Type    string
	Prefix  string
	Color   string
	Message string
	Lines   []string
}

var (
	UI   *TerminalUI
	once sync.Once
)

func Init() {
	once.Do(func() {
		UI = &TerminalUI{
			eventQueue: make(chan UIEvent, 1000),
		}
		UI.StartLoop()
	})
}

func (ui *TerminalUI) CleanProgress() {
	if ui.progressSize > 0 {
		for range ui.progressSize {
			fmt.Print("\033[1A")
			fmt.Print("\033[2K")
		}
	}
	ui.progressSize = 0
}

func (ui *TerminalUI) LogLine(prefix, color, msg string) {
	ui.eventQueue <- UIEvent{Type: "log", Prefix: prefix, Color: color, Message: msg}
}

func (ui *TerminalUI) SetProgress(lines ...string) {
	ui.eventQueue <- UIEvent{Type: "progress", Lines: lines}
}

func (ui *TerminalUI) StartLoop() {
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		var pendingLogs []UIEvent
		var pendingProgress []string

		for {
			select {
			case ev := <-ui.eventQueue:
				if ev.Type == "log" {
					pendingLogs = append(pendingLogs, ev)
				} else if ev.Type == "progress" {
					pendingProgress = ev.Lines
				}
			case <-ticker.C:
				ui.mu.Lock()
				ui.CleanProgress()

				for _, log := range pendingLogs {
					entry := fmt.Sprintf("%s%s%s\033[0m", log.Color, log.Prefix, log.Message)
					fmt.Fprintln(os.Stdout, entry)
				}
				pendingLogs = nil

				for _, line := range pendingProgress {
					fmt.Fprintln(os.Stdout, line)
				}
				ui.progressSize = len(pendingProgress)
				ui.mu.Unlock()
			}
		}
	}()
}
