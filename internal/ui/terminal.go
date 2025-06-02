package ui

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

type TerminalUI struct {
	mu         sync.Mutex
	eventQueue chan UIEvent
	bottonSize int
	done       chan struct{}
	wg         sync.WaitGroup
}

type UIEvent struct {
	Type     string
	Prefix   string
	Color    string
	Message  string
	Total    int
	Progress ProgressEvent
}

func (ui *TerminalUI) CleanBotton() {
	if ui.bottonSize == 0 {
		return
	}

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("\033[%dA", ui.bottonSize))

	for range ui.bottonSize {
		sb.WriteString("\033[2K")
		sb.WriteString("\033[1B")
	}

	sb.WriteString(fmt.Sprintf("\033[%dA", ui.bottonSize))

	fmt.Print(sb.String())
}

func (ui *TerminalUI) SendLogEvent(prefix, color, msg string, fixed bool) {
	if fixed {
		ui.sendEvent(UIEvent{Type: "fixed", Prefix: prefix, Color: color, Message: msg})
	} else {
		ui.sendEvent(UIEvent{Type: "log", Prefix: prefix, Color: color, Message: msg})
	}
}

type ProgressEvent int

const (
	Success ProgressEvent = iota
	Error
	Retry
	Inc
	Tor
	Total
)

func (ui *TerminalUI) SendProgressEvent(event ProgressEvent) {
	ui.sendEvent(UIEvent{Type: "event", Progress: event})
}

func (ui *TerminalUI) SendIpProgressEvent(ip string) {
	ui.sendEvent(UIEvent{Type: "event", Progress: Tor, Message: ip})
}

func (ui *TerminalUI) SendTotalProgressEvent(total int) {
	ui.sendEvent(UIEvent{Type: "event", Progress: Total, Total: total})
}

func (ui *TerminalUI) sendEvent(event UIEvent) {
	ui.eventQueue <- event
}

func (ui *TerminalUI) StartLoop() {
	ui.wg.Add(1)

	go func() {
		defer ui.wg.Done()

		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		var pendingLogs []UIEvent
		var fixed []UIEvent

		for {
			select {
			case ev := <-ui.eventQueue:
				if ev.Type == "log" {
					pendingLogs = append(pendingLogs, ev)
				} else if ev.Type == "fixed" {
					fixed = append(fixed, ev)
				} else if ev.Type == "event" {
					switch ev.Progress {
					case Success:
						progress.AddSuccess()
						break
					case Error:
						progress.AddError()
						break
					case Retry:
						progress.AddRetry()
						break
					case Inc:
						progress.Inc()
						break
					case Tor:
						progress.SetTor(ev.Message)
					case Total:
						progress.SetTotal(ev.Total)
					}
				}
			case <-ticker.C:
				ui.mu.Lock()
				ui.CleanBotton()

				ui.bottonSize = 0

				var outputBuffer bytes.Buffer

				for _, log := range pendingLogs {
					entry := fmt.Sprintf("%s%s%s\033[0m\n", log.Color, log.Prefix, log.Message)
					outputBuffer.WriteString(entry)
				}
				pendingLogs = nil

				for _, log := range fixed {
					entry := fmt.Sprintf("%s%s%s\033[0m\n", log.Color, log.Prefix, log.Message)
					outputBuffer.WriteString(entry)
				}
				ui.bottonSize += len(fixed)

				pendingProgress := progress.render()
				for _, line := range pendingProgress {
					outputBuffer.WriteString(line + "\n")
				}

				fmt.Fprint(os.Stdout, outputBuffer.String())

				ui.bottonSize += len(pendingProgress)
				ui.mu.Unlock()
			case <-ui.done:
				if pendingLogs == nil {
					return
				}
			}
		}
	}()
}

func (ui *TerminalUI) Stop() {
	close(ui.done)
	ui.wg.Wait()
}
