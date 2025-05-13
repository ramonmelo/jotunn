package logger

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/LinharesAron/jotunn/internal/ui"
)

type ProgressTracker struct {
	total         int
	current       int
	success       int
	errors        int
	retries       int
	startTime     time.Time
	lastETAUpdate time.Time
	lastETA       time.Duration
	isTor         bool
	torIP         string
	mu            sync.Mutex
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

func (p *ProgressTracker) AddSuccess() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.success++
}

func (p *ProgressTracker) AddError() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.errors++
}

func (p *ProgressTracker) AddRetry() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.retries++
}

func (p *ProgressTracker) SetTor(ip string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.isTor = true
	p.torIP = ip
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

	elapsed := time.Since(startTime)
	rate := float64(current) / elapsed.Seconds()
	etaStr := "--"
	now := time.Now()

	if elapsed >= 10*time.Second && rate > 0 {
		rawETA := time.Duration(float64(p.total-current)/rate) * time.Second

		if p.lastETAUpdate.IsZero() || now.Sub(p.lastETAUpdate) >= 50*time.Second {
			p.lastETA = rawETA
			p.lastETAUpdate = now
		}

		etaStr = formatDuration(p.lastETA)
	}

	tor := ""
	if p.isTor {
		tor = fmt.Sprintf(" 🧅 %s", p.torIP)
	}

	status := fmt.Sprintf(
		"🎯 %d ❌ %d 🔁 %d%s",
		p.success,
		p.errors,
		p.retries,
		tor,
	)

	draw := fmt.Sprintf(
		"🔥 Bruteforcing %s %d/%d (%.1f it/s) ⏱ %s ⌛ %s",
		bar,
		current, total,
		rate,
		formatDuration(elapsed),
		etaStr,
	)

	ui.UI.SetProgress(status, draw)
}

func formatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60

	if h > 0 {
		return fmt.Sprintf("%02dh%02dm%02ds", h, m, s)
	}
	return fmt.Sprintf("%02dm%02ds", m, s)
}
