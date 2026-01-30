package utils

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"os"
)

type ExtraDataAfterJSONError struct{}

func (e *ExtraDataAfterJSONError) Error() string {
	return "extra data after JSON object"
}

// FromJSON decodes JSON from byte slice (wrapper around streaming version).
//
//nolint:ireturn // Generic functions must return type parameter T
func FromJSON[T any](data []byte) (T, error) {
	var result T
	if len(data) == 0 {
		return result, nil
	}

	result, err := FromJSONStream[T](bytes.NewReader(data))

	return result, err
}

// FromJSONStream decodes JSON from io.Reader (streaming version).
//
//nolint:ireturn // Generic functions must return type parameter T
func FromJSONStream[T any](r io.Reader) (T, error) {
	var result T

	decoder := json.NewDecoder(r)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(&result)
	if err != nil {
		return result, err
	}

	if decoder.More() {
		return result, &ExtraDataAfterJSONError{}
	}

	return result, nil
}

// MustFromJSON decodes JSON from byte slice (wrapper around streaming version).
//
//nolint:ireturn // Generic functions must return type parameter T
func MustFromJSON[T any](data []byte) T {
	result, err := FromJSON[T](data)
	if err != nil {
		slog.Error("failed to unmarshal JSON", ErrAttr(err))
		os.Exit(1)
	}

	return result
}

// newJsonEncoder creates a new JSON encoder with the default settings.
func newJsonEncoder(w io.Writer) *json.Encoder {
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "")

	return encoder
}

// ToJSON encodes to JSON byte slice (wrapper around streaming version).
func ToJSON(v any) ([]byte, error) {
	var buf bytes.Buffer
	if err := ToJSONStream(&buf, v); err != nil {
		return nil, err
	}

	return bytes.TrimSpace(buf.Bytes()), nil
}

// ToJSONIndent encodes to indented JSON byte slice (wrapper around streaming version).
func ToJSONIndent(v any) ([]byte, error) {
	var buf bytes.Buffer
	if err := ToJSONStreamIndent(&buf, v); err != nil {
		return nil, err
	}

	return bytes.TrimSpace(buf.Bytes()), nil
}

// ToJSONStream encodes to JSON and writes to io.Writer (streaming version).
func ToJSONStream(w io.Writer, v any) error {
	encoder := newJsonEncoder(w)

	return encoder.Encode(v)
}

// ToJSONStreamIndent encodes to indented JSON and writes to io.Writer (streaming version).
func ToJSONStreamIndent(w io.Writer, v any) error {
	encoder := newJsonEncoder(w)
	encoder.SetIndent("", "  ")

	return encoder.Encode(v)
}

func MustToJSON(v any) []byte {
	data, err := ToJSON(v)
	if err != nil {
		slog.Error("failed to marshal JSON", ErrAttr(err))
		os.Exit(1)
	}

	return data
}

func MustToJSONIndent(v any) []byte {
	data, err := ToJSONIndent(v)
	if err != nil {
		slog.Error("failed to marshal JSON", ErrAttr(err))
		os.Exit(1)
	}

	return data
}
