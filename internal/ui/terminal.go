package ui

import (
	"fmt"
	"os"
	"sync"
)

type TerminalUI struct {
	mu           sync.Mutex
	progressSize int
}

var UI *TerminalUI

func Init() {
	UI = &TerminalUI{}
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
	ui.mu.Lock()
	defer ui.mu.Unlock()
	ui.CleanProgress()

	entry := fmt.Sprintf("%s%s%s\033[0m", color, prefix, msg)
	fmt.Fprintln(os.Stdout, entry)
}

func (ui *TerminalUI) SetProgress(lines ...string) {
	ui.mu.Lock()
	defer ui.mu.Unlock()
	ui.CleanProgress()

	for _, line := range lines {
		fmt.Fprintln(os.Stdout, line)
	}

	ui.progressSize = len(lines)
}
