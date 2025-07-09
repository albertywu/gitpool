package internal

import (
	"fmt"
	"log"
	"os"
	"text/tabwriter"
	"time"
)

// InitLogger initializes the logger with a custom format
func InitLogger() {
	log.SetFlags(0) // Remove default timestamp
	log.SetPrefix("")
}

// PrintError prints an error message in a formatted way
func PrintError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "[ERROR] "+format+"\n", args...)
}

// PrintInfo prints an info message in a formatted way
func PrintInfo(format string, args ...interface{}) {
	fmt.Printf("[INFO] "+format+"\n", args...)
}

// PrintWarn prints a warning message in a formatted way
func PrintWarn(format string, args ...interface{}) {
	fmt.Printf("[WARN] "+format+"\n", args...)
}

// NewTabWriter creates a new tab writer for formatted output
func NewTabWriter() *tabwriter.Writer {
	return tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
}

// FormatDuration formats a duration for display
func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	} else if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	} else if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	return fmt.Sprintf("%dd", int(d.Hours()/24))
}

// FormatTime formats a time for display
func FormatTime(t *time.Time) string {
	if t == nil {
		return "Never"
	}
	return t.Format("2006-01-02 15:04:05")
}
