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
		return 80, 24
	}
	return width, height
}

func clearAndPosition() {
	fmt.Print("\033[2J\033[H")
}

func drawFullScreen(content []string, footer string) {
	width, height := getTerminalSize()
	clearAndPosition()

	linesPrinted := 0
	for _, line := range content {
		if linesPrinted >= height-2 {
			break
		}
		if len(line) > width {
			line = line[:width]
		}
		fmt.Println(line)
		linesPrinted++
	}

	for linesPrinted < height-2 {
		fmt.Println()
		linesPrinted++
	}

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
