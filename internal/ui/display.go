package ui

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

// Colors for terminal output
const (
	ColorReset  = "\033[0m"
	ColorGreen  = "\033[0;32m"
	ColorYellow = "\033[1;33m"
	ColorBlue   = "\033[0;34m"
	ColorRed    = "\033[0;31m"
)

// PrintColored prints text with specified color
func PrintColored(color, format string, args ...interface{}) {
	fmt.Printf(color+format+ColorReset, args...)
}

// PrintInfo logs an informational message both to the console and to the logger
func PrintInfo(logger *slog.Logger, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	fmt.Println(message)
	if logger != nil {
		logger.Info(message)
	}
}

// PrintError logs an error message both to the console and to the logger
func PrintError(logger *slog.Logger, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	PrintColored(ColorRed, "ERROR: %s\n", message)
	if logger != nil {
		logger.Error(message)
	}
}

// PrintWarning logs a warning message both to the console and to the logger
func PrintWarning(logger *slog.Logger, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	PrintColored(ColorYellow, "WARNING: %s\n", message)
	if logger != nil {
		logger.Warn(message)
	}
}

// DisplayMenu shows the main menu and returns the selected option
func DisplayMenu() int {
	fmt.Println()
	PrintColored(ColorBlue, "Select operation:\n")
	fmt.Println("1) Analyze only - show what changes would be made (no changes)")
	fmt.Println("2) Sync files - copy only new/changed files (no deletions)")
	fmt.Println("3) Full sync - copy new/changed files AND delete removed files")
	fmt.Println("4) Exit")
	fmt.Println()

	var option int
	fmt.Print("Enter option (1-4): ")
	fmt.Scanln(&option)
	fmt.Println()

	return option
}

// ConfirmAction asks for user confirmation with the given prompt
func ConfirmAction(prompt string) bool {
	fmt.Print(prompt + " (y/n) ")
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}

// ConfirmDangerous asks for strict confirmation for dangerous operations
func ConfirmDangerous(prompt string) bool {
	fmt.Print(prompt + " Type 'yes' to confirm: ")
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(response)
	return response == "yes"
}

// DisplayConfig shows the current configuration
func DisplayConfig(source, target, logFile string, includeDirs, excludePatterns []string) {
	PrintColored(ColorBlue, "Media Backup - Configuration\n")
	PrintColored(ColorYellow, "Source: ")
	fmt.Println(source)
	PrintColored(ColorYellow, "Target: ")
	fmt.Println(target)
	PrintColored(ColorYellow, "Log: ")
	fmt.Println(logFile)
	fmt.Println()

	// Display include directories
	PrintColored(ColorYellow, "Include Directories:\n")
	if len(includeDirs) > 0 {
		for _, dir := range includeDirs {
			fmt.Printf("  - %s\n", dir)
		}
	} else {
		fmt.Println("  (All directories)")
	}
	fmt.Println()

	// Display exclude patterns
	PrintColored(ColorYellow, "Exclude Patterns:\n")
	if len(excludePatterns) > 0 {
		for _, pattern := range excludePatterns {
			fmt.Printf("  - %s\n", pattern)
		}
	} else {
		fmt.Println("  (None)")
	}
	fmt.Println()
}

// FormatDuration formats a duration in seconds to a human-readable string
func FormatDuration(seconds int) string {
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	secs := seconds % 60

	return fmt.Sprintf("%dh %dm %ds", hours, minutes, secs)
}

// UpdateProgress displays the current progress percentage
func UpdateProgress(percent int) {
	fmt.Printf("\rProgress: %d%%", percent)
	if percent >= 100 {
		fmt.Println()
	}
}
