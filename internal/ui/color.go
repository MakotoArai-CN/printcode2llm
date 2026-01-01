package ui

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

var (
	colorCyan   = color.New(color.FgCyan, color.Bold)
	colorGreen  = color.New(color.FgGreen, color.Bold)
	colorRed    = color.New(color.FgRed, color.Bold)
	colorYellow = color.New(color.FgYellow)
	colorWhite  = color.New(color.FgWhite)
	colorBlue   = color.New(color.FgBlue, color.Bold)
)

func PrintHeader(text string) {
	fmt.Println()
	line := strings.Repeat("─", 50)
	colorCyan.Println("┌" + line + "┐")

	textLen := len([]rune(text))
	padding := (50 - textLen) / 2
	if padding < 0 {
		padding = 0
	}

	colorCyan.Print("│")
	fmt.Print(strings.Repeat(" ", padding))
	colorCyan.Print(text)
	fmt.Print(strings.Repeat(" ", 50-padding-textLen*2))
	colorCyan.Println("│")

	colorCyan.Println("└" + line + "┘")
	fmt.Println()
}

func PrintSection(format string, args ...interface{}) {
	colorBlue.Printf("▶ "+format+"\n", args...)
}

func PrintInfo(format string, args ...interface{}) {
	colorCyan.Printf("  ℹ "+format+"\n", args...)
}

func PrintSuccess(format string, args ...interface{}) {
	colorGreen.Printf("  ✓ "+format+"\n", args...)
}

func PrintWarning(format string, args ...interface{}) {
	colorYellow.Printf("  ⚠ "+format+"\n", args...)
}

func PrintError(format string, args ...interface{}) {
	colorRed.Printf("  ✗ "+format+"\n", args...)
}

func PrintStep(format string, args ...interface{}) {
	colorWhite.Printf("  → "+format+"\n", args...)
}

func FormatNumber(n int) string {
	str := fmt.Sprintf("%d", n)
	if len(str) <= 3 {
		return str
	}

	var result []byte
	for i, ch := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, byte(ch))
	}
	return string(result)
}

func FormatBytes(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}

	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}