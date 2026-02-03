package types

import (
	"bytes"
	"fmt"
	"http-mqtt-boilerplate/backend/pkg/utils"
	"net/url"
)

// URL wraps net/url.URL and marshals as a string instead of an object.
//
//nolint:recvcheck // Mixed receivers required: UnmarshalJSON needs pointer (mutates), MarshalJSON/String need value (non-addressable support)
type URL struct {
	*url.URL
}

// NewURL creates a new URL from a string.
func NewURL(s string) (URL, error) {
	u, err := url.Parse(s)
	if err != nil {
		return URL{}, err
	}

	return URL{URL: u}, nil
}

// MustNewURL creates a new URL from a string and panics on error.
func MustNewURL(s string) URL {
	u, err := NewURL(s)
	if err != nil {
		panic(err)
	}

	return u
}

// MarshalJSON marshals the URL as a JSON string.
// Value receiver ensures this works for non-addressable values (map values, interface{}, etc).
func (u URL) MarshalJSON() ([]byte, error) {
	if u.URL == nil {
		return utils.ToJSON("")
	}

	return utils.ToJSON(u.String())
}

// UnmarshalJSON unmarshals a JSON string into a URL.
func (u *URL) UnmarshalJSON(data []byte) error {
	// Handle JSON null explicitly
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		u.URL = nil

		return nil
	}

	s, err := utils.FromJSON[string](data)
	if err != nil {
		return err
	}

	if s == "" {
		u.URL = nil

		return nil
	}

	parsed, err := url.Parse(s)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	u.URL = parsed

	return nil
}

// String returns the string representation of the URL.
// Value receiver ensures this works for non-addressable values (map values, interface{}, etc).
func (u URL) String() string {
	if u.URL == nil {
		return ""
	}

	return u.URL.String()
}
