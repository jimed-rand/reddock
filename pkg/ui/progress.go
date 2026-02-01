package ui

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// Progress handles displaying progress bars and spinners
type Progress struct {
	message    string
	total      int
	current    int
	isFinished bool
	mu         sync.Mutex
	stopChan   chan bool
	isSpinner  bool

	// Spinner chars
	spinnerChars []string
}

// NewProgressBar creates a determinate progress bar
func NewProgressBar(total int, message string) *Progress {
	return &Progress{
		total:     total,
		message:   message,
		isSpinner: false,
	}
}

// NewSpinner creates an indeterminate spinner
func NewSpinner(message string) *Progress {
	return &Progress{
		message:      message,
		isSpinner:    true,
		stopChan:     make(chan bool),
		spinnerChars: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
	}
}

// Start begins the progress display
func (p *Progress) Start() {
	if p.isSpinner {
		go p.spin()
	} else {
		p.render()
	}
}

func (p *Progress) spin() {
	i := 0
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-p.stopChan:
			return
		case <-ticker.C:
			p.mu.Lock()
			char := p.spinnerChars[i%len(p.spinnerChars)]
			fmt.Printf("\r\033[K%s %s", char, p.message)
			i++
			p.mu.Unlock()
		}
	}
}

// Increment increases progress by 1 (for progress bar)
func (p *Progress) Increment() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.current++
	p.render()
}

// Update sets the current progress (for progress bar)
func (p *Progress) Update(current int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.current = current
	p.render()
}

// SetMessage updates the message displayed
func (p *Progress) SetMessage(msg string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.message = msg
	if !p.isSpinner {
		p.render()
	}
}

func (p *Progress) render() {
	if p.isFinished {
		return
	}
	width := 40
	percent := float64(p.current) / float64(p.total) * 100
	if percent > 100 {
		percent = 100
	}

	filled := int(float64(width) * (float64(p.current) / float64(p.total)))
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("-", width-filled)
	fmt.Printf("\r\033[K%s [%s] %.0f%%", p.message, bar, percent)
}

// Finish completes the progress
func (p *Progress) Finish(finalMsg string) {
	p.mu.Lock()
	// No defer unlock here because we might need to print after
	p.isFinished = true
	if p.isSpinner {
		close(p.stopChan)
		// Wait a tiny bit to ensure spinner goroutine stops printing
		time.Sleep(10 * time.Millisecond)
		// Clear line and print final status
		fmt.Printf("\r\033[K✔ %s\n", finalMsg)
	} else {
		p.current = p.total
		// Temporarily unlock to call render (which locks) - actually render doesn't lock here because we are inside lock?
		// Wait, render() does NOT lock. It assumes caller holds lock if needed?
		// My helper methods like Increment() lock. render() does not lock.
		// So I CAN call render() here.

		// But wait, render uses p.current/p.total.
		// I should check if render uses any shared resource that needs protection?
		// fmt.Printf is generally thread-safe for the writer, but mixed writes on line might conflict.

		width := 40
		bar := strings.Repeat("█", width)
		fmt.Printf("\r\033[K%s [%s] 100%%", p.message, bar)
		fmt.Println()
		if finalMsg != "" {
			fmt.Println(finalMsg)
		}
	}
	p.mu.Unlock()
}

// IsDone checks if the progress is finished
func (p *Progress) IsDone() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.isFinished
}
