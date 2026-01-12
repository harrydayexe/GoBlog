package logger

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"

	"github.com/fatih/color"
)

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

type CLIHandler struct {
	slog.Handler
	opts *slog.HandlerOptions
	l    *log.Logger
}

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
