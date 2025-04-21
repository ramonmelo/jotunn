package logger

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

type ProgressTracker struct {
	total     int
	current   int
	startTime time.Time
	mu        sync.Mutex
}

var Progress *ProgressTracker

func InitProgressTracker(total int) {
	Progress = &ProgressTracker{
		total:     total,
		startTime: time.Now(),
	}
}

func (p *ProgressTracker) Inc() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.current++
	current := p.current
	total := p.total
	startTime := p.startTime

	p.render(current, total, startTime)
}

func (p *ProgressTracker) renderInline() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.current < p.total {
		p.render(p.current, p.total, p.startTime)
	}
}

func (p *ProgressTracker) render(current, total int, startTime time.Time) {
	width := 30
	progress := float64(current) / float64(total)
	filled := int(progress * float64(width))

	bar := "[" + strings.Repeat("=", filled) + ">" + strings.Repeat(" ", width-filled) + "]"

	elapsed := time.Since(startTime).Seconds()
	rate := float64(current) / elapsed

	fmt.Fprintf(os.Stdout, "\r🔥 Bruteforcing %s %d/%d (%.1f it/s)", bar, current, total, rate)

	if current == total {
		fmt.Println()
	}
}
