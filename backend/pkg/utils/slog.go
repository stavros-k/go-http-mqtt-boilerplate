package utils

import (
	"bytes"
	"log/slog"
)

// logWriter is a small io.Writer that writes to a slog.Logger.
type logWriter struct {
	logger *slog.Logger
}

func NewSlogWriter(logger *slog.Logger) *logWriter {
	return &logWriter{logger: logger}
}

// Write implements io.Writer by logging each line at Info level.
func (w *logWriter) Write(p []byte) (n int, err error) {
	// Trim trailing newline to avoid double newlines in logs
	msg := bytes.TrimRight(p, "\n")
	if len(msg) > 0 {
		w.logger.Info(string(msg))
	}

	return len(p), nil
}

func ErrAttr(err error) slog.Attr {
	return slog.Any("error", err)
}

func SlogReplacer(groups []string, a slog.Attr) slog.Attr {
	timeFormat := "2006-01-02 15:04:05"

	//nolint:exhaustive // Only formatting specific types; other types pass through unchanged
	switch a.Value.Kind() {
	case slog.KindTime:
		a.Value = slog.StringValue(a.Value.Time().Format(timeFormat))
	case slog.KindDuration:
		a.Value = slog.StringValue(a.Value.Duration().String())
	}

	return a
}
