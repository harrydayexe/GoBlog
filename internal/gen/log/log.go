package log

import (
	"fmt"
	"os"
)

// CLILogger is a simple logger for command line applications that outputs
// colored logs
type CLILogger struct {
	verbose bool
	width   string
	key     string
}

// Info outputs an info log to the stdout
func (l CLILogger) Info(key string, format string, args ...any) {
	p := fmt.Sprintf("\033[36m%"+l.width+"s \033[90m:\033[0m ", key)
	fmt.Printf(p+format+"\n", args...)
}

// Debug outputs a debug log to the stdout if verbose mode is enabled
func (l CLILogger) Debug(key string, format string, args ...any) {
	if l.verbose {
		p := fmt.Sprintf("\033[96m%"+l.width+"s \033[90m:\033[0m ", key)
		fmt.Printf(p+format+"\n", args...)
	}
}

// Warn outputs a warning log to the stdout
func (l CLILogger) Warn(format string, args ...any) {
	p := fmt.Sprintf("\033[33m%"+l.width+"s \033[90m:\033[0m ", "warn")
	fmt.Printf(p+format+"\n", args...)
}

// Error outputs an error log to the stderr
func (l CLILogger) Error(err error) {
	p := fmt.Sprintf("\033[31m%"+l.width+"s \033[90m:\033[0m ", "error")
	fmt.Fprintf(os.Stderr, p+"%s\n", err.Error())
}

// NewCLILogger creates a new instance of CLILogger with the specified key,
// verbosity and column width.
func NewCLILogger(key string, verbose bool, width int) *CLILogger {
	return &CLILogger{
		verbose: verbose,
		width:   fmt.Sprintf("%d", width),
		key:     key,
	}
}
