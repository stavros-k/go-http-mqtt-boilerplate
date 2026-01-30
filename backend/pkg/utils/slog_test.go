package utils

import (
	"bytes"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"
)

func TestNewSlogWriter(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	writer := NewSlogWriter(logger)
	if writer == nil {
		t.Fatal("NewSlogWriter() returned nil")
	}

	if writer.logger != logger {
		t.Error("NewSlogWriter() did not set logger correctly")
	}
}

func TestLogWriter_Write(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		wantLogs bool
	}{
		{
			name:     "simple message",
			input:    "test message\n",
			wantLogs: true,
		},
		{
			name:     "message without newline",
			input:    "test message",
			wantLogs: true,
		},
		{
			name:     "empty message",
			input:    "",
			wantLogs: false,
		},
		{
			name:     "only newline",
			input:    "\n",
			wantLogs: false,
		},
		{
			name:     "multiple newlines",
			input:    "test\n\n",
			wantLogs: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			logger := slog.New(slog.NewTextHandler(&buf, nil))
			writer := NewSlogWriter(logger)

			n, err := writer.Write([]byte(tt.input))
			if err != nil {
				t.Fatalf("Write() error = %v", err)
			}

			if n != len(tt.input) {
				t.Errorf("Write() returned n = %v, want %v", n, len(tt.input))
			}

			output := buf.String()
			if tt.wantLogs && output == "" {
				t.Error("Write() should have logged but didn't")
			}

			if !tt.wantLogs && output != "" {
				t.Errorf("Write() should not have logged but did: %v", output)
			}
		})
	}
}

func TestLogWriter_WriteMultipleLines(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))
	writer := NewSlogWriter(logger)

	messages := []string{"line 1\n", "line 2\n", "line 3\n"}

	for _, msg := range messages {
		_, err := writer.Write([]byte(msg))
		if err != nil {
			t.Fatalf("Write() error = %v", err)
		}
	}

	output := buf.String()
	for _, msg := range messages {
		trimmed := strings.TrimRight(msg, "\n")
		if !strings.Contains(output, trimmed) {
			t.Errorf("Output should contain %q, but got %q", trimmed, output)
		}
	}
}

func TestErrAttr(t *testing.T) {
	t.Parallel()

	err := errors.New("test error")
	attr := ErrAttr(err)

	if attr.Key != "error" {
		t.Errorf("ErrAttr() Key = %v, want %v", attr.Key, "error")
	}

	if attr.Value.Any() != err {
		t.Errorf("ErrAttr() Value = %v, want %v", attr.Value.Any(), err)
	}
}

func TestErrAttrNil(t *testing.T) {
	t.Parallel()

	attr := ErrAttr(nil)

	if attr.Key != "error" {
		t.Errorf("ErrAttr() Key = %v, want %v", attr.Key, "error")
	}

	if attr.Value.Any() != nil {
		t.Errorf("ErrAttr() Value should be nil, got %v", attr.Value.Any())
	}
}

func TestSlogReplacer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		attr   slog.Attr
		verify func(t *testing.T, result slog.Attr)
	}{
		{
			name: "time attribute",
			attr: slog.Time("timestamp", time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC)),
			verify: func(t *testing.T, result slog.Attr) {
				if result.Value.Kind() != slog.KindString {
					t.Errorf("Time should be converted to string, got %v", result.Value.Kind())
				}
				expected := "2024-01-15 10:30:45"
				if result.Value.String() != expected {
					t.Errorf("Time format = %v, want %v", result.Value.String(), expected)
				}
			},
		},
		{
			name: "duration attribute",
			attr: slog.Duration("elapsed", 5*time.Second+250*time.Millisecond),
			verify: func(t *testing.T, result slog.Attr) {
				if result.Value.Kind() != slog.KindString {
					t.Errorf("Duration should be converted to string, got %v", result.Value.Kind())
				}
				if result.Value.String() != "5.25s" {
					t.Errorf("Duration format = %v, want %v", result.Value.String(), "5.25s")
				}
			},
		},
		{
			name: "string attribute unchanged",
			attr: slog.String("name", "test"),
			verify: func(t *testing.T, result slog.Attr) {
				if result.Value.Kind() != slog.KindString {
					t.Errorf("String kind should be preserved, got %v", result.Value.Kind())
				}
				if result.Value.String() != "test" {
					t.Errorf("String value = %v, want %v", result.Value.String(), "test")
				}
			},
		},
		{
			name: "int attribute unchanged",
			attr: slog.Int("count", 42),
			verify: func(t *testing.T, result slog.Attr) {
				if result.Value.Kind() != slog.KindInt64 {
					t.Errorf("Int kind should be preserved, got %v", result.Value.Kind())
				}
				if result.Value.Int64() != 42 {
					t.Errorf("Int value = %v, want %v", result.Value.Int64(), 42)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := SlogReplacer(nil, tt.attr)
			tt.verify(t, result)
		})
	}
}

func TestLogOnError(t *testing.T) {
	t.Parallel()

	t.Run("no error", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&buf, nil))

		LogOnError(logger, func() error {
			return nil
		}, "test message")

		if buf.Len() > 0 {
			t.Errorf("LogOnError() should not log when no error, got: %s", buf.String())
		}
	})

	t.Run("with error", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&buf, nil))

		testErr := errors.New("test error")
		LogOnError(logger, func() error {
			return testErr
		}, "operation failed")

		output := buf.String()
		if !strings.Contains(output, "operation failed") {
			t.Errorf("LogOnError() should contain message, got: %s", output)
		}

		if !strings.Contains(output, "test error") {
			t.Errorf("LogOnError() should contain error, got: %s", output)
		}
	})

	t.Run("panic in function", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&buf, nil))

		defer func() {
			if r := recover(); r == nil {
				t.Error("LogOnError() should propagate panic")
			}
		}()

		LogOnError(logger, func() error {
			panic("test panic")
		}, "should panic")
	})
}

func TestLogWriter_WriteReturnValue(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))
	writer := NewSlogWriter(logger)

	data := []byte("test data")
	n, err := writer.Write(data)

	if err != nil {
		t.Fatalf("Write() returned unexpected error: %v", err)
	}

	if n != len(data) {
		t.Errorf("Write() returned n = %v, want %v", n, len(data))
	}
}