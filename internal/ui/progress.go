package ui

import (
	"fmt"
	"strings"
	"sync"
	"time"
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
	smoothedRate  float64
}

func (p *ProgressTracker) SetTotal(total int) {
	p.total = total
}

func (p *ProgressTracker) Inc() {
	p.current++
}

func (p *ProgressTracker) AddSuccess() {
	p.success++
}

func (p *ProgressTracker) AddError() {
	p.errors++
}

func (p *ProgressTracker) AddRetry() {
	p.retries++
}

func (p *ProgressTracker) SetTor(ip string) {
	p.isTor = true
	p.torIP = ip
}

func (p *ProgressTracker) render() []string {
	if p.total == 0 {
		return []string{}
	}

	width := 30
	progress := float64(p.current) / float64(p.total)
	filled := int(progress * float64(width))
	bar := "[" + strings.Repeat("=", filled) + ">" + strings.Repeat(" ", width-filled) + "]"

	elapsed := time.Since(p.startTime)
	rate := float64(p.current) / elapsed.Seconds()
	etaStr := "--"
	now := time.Now()

	if elapsed >= 10*time.Second && rate > 0 {
		rawETA := time.Duration(float64(p.total-p.current)/rate) * time.Second

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
		p.current, p.total,
		rate,
		formatDuration(elapsed),
		etaStr,
	)

	return []string{status, draw}
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
