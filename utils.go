package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/term"
)

func getTerminalSize() (width, height int) {
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 80, 24 // fallback defaults
	}
	return width, height
}

func clearAndPosition() {
	fmt.Print("\033[2J\033[H") // clear screen and move cursor to top-left
}

func drawFullScreen(content []string, footer string) {
	width, height := getTerminalSize()
	clearAndPosition()

	// Draw content lines
	linesPrinted := 0
	for _, line := range content {
		if linesPrinted >= height-2 {
			break
		}
		// Truncate if too wide
		if len(line) > width {
			line = line[:width]
		}
		fmt.Println(line)
		linesPrinted++
	}

	// Fill remaining space
	for linesPrinted < height-2 {
		fmt.Println()
		linesPrinted++
	}

	// Draw separator and footer at bottom
	fmt.Println(strings.Repeat("-", width))
	if len(footer) > width {
		footer = footer[:width]
	}
	fmt.Print(footer)
}

func formatDuration(nanoseconds int64) string {
	if nanoseconds == 0 {
		return "0s"
	}

	duration := time.Duration(nanoseconds)
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	seconds := int(duration.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}
