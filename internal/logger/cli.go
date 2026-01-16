// Package logger provides CLI logging with colored output using slog.
package logger

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"

	"github.com/fatih/color"
)

// Log represents a formatted log message with color and level information.
type Log struct {
	ColorFunc  func(format string, a ...any) string
	Level      string
	Message    string
	Attributes string
}

func (l Log) String() string {
	if l.Attributes != "" {
		return fmt.Sprintf("%s%s\n%s", l.ColorFunc(l.Level), l.ColorFunc(l.Message), l.ColorFunc(l.Attributes))
	} else {
		return fmt.Sprintf("%s%s", l.ColorFunc(l.Level), l.ColorFunc(l.Message))
	}
}

func newLog() Log {
	return Log{
		ColorFunc:  color.WhiteString,
		Level:      "",
		Message:    "",
		Attributes: "",
	}
}

// CLIHandler implements slog.Handler with colored output for CLI applications.
type CLIHandler struct {
	slog.Handler
	opts *slog.HandlerOptions
	l    *log.Logger
}

// NewCLIHandlerWithOptions creates a CLIHandler with the specified handler options.
func NewCLIHandlerWithOptions(
	out io.Writer,
	opts slog.HandlerOptions,
) *CLIHandler {
	h := &CLIHandler{
		Handler: slog.NewTextHandler(out, &opts),
		l:       log.New(out, "", 0),
	}

	return h
}

// NewDefaultCLIHandler creates a CLIHandler with default options.
func NewDefaultCLIHandler(
	out io.Writer,
) *CLIHandler {
	opts := &slog.HandlerOptions{}
	h := &CLIHandler{
		Handler: slog.NewTextHandler(out, opts),
		opts:    opts,
		l:       log.New(out, "", 0),
	}

	return h
}

// NewDefaultCLIHandlerWithVerbosity creates a CLIHandler with a specific verbosity level.
func NewDefaultCLIHandlerWithVerbosity(
	out io.Writer,
	verboseLevel slog.Level,
) *CLIHandler {
	opts := &slog.HandlerOptions{
		Level: verboseLevel,
	}
	h := &CLIHandler{
		Handler: slog.NewTextHandler(out, opts),
		opts:    opts,
		l:       log.New(out, "", 0),
	}

	return h
}

// Handle implements slog.Handler by formatting and outputting log records with colors.
func (h *CLIHandler) Handle(ctx context.Context, r slog.Record) error {
	loggingLevel := h.opts.Level.Level()

	log := newLog()
	setLevel(r, &log, loggingLevel < slog.LevelInfo)
	log.Message = r.Message

	if loggingLevel < slog.LevelInfo {
		processAttributes(r, &log)
	}

	h.l.Println(log)
	return nil
}

func setLevel(r slog.Record, log *Log, showInfo bool) {
	log.Level = r.Level.String() + ": "

	switch r.Level {
	case slog.LevelDebug:
		log.ColorFunc = color.CyanString
	case slog.LevelInfo:
		if !showInfo {
			log.Level = ""
		}
		log.ColorFunc = color.WhiteString
	case slog.LevelWarn:
		log.ColorFunc = color.YellowString
	case slog.LevelError:
		log.ColorFunc = color.RedString
	}
}

func processAttributes(r slog.Record, log *Log) {
	fields := make(map[string]string, r.NumAttrs())
	r.Attrs(func(a slog.Attr) bool {
		fields[a.Key] = a.Value.String()

		return true
	})

	if len(fields) == 0 {
		return
	}

	log.Attributes = "Attributes:"

	for k, v := range fields {
		log.Attributes = log.Attributes + fmt.Sprintf("\n - %s: %s", k, v)
	}
}
