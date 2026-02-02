package generate

import (
	"reflect"
	"testing"
)

func TestExtractParamName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		path     string
		expected []string
		wantErr  bool
	}{
		{
			name:     "no params",
			path:     "/users",
			expected: []string{},
		},
		{
			name:     "simple one param",
			path:     "/users/{userID}",
			expected: []string{"userID"},
		},
		{
			name:     "simple one param with regex",
			path:     "/users/{userID:[0-9]+}",
			expected: []string{"userID"},
		},
		{
			name:     "multiple params",
			path:     "/users/{userID}/{userName}",
			expected: []string{"userID", "userName"},
		},
		{
			name:     "multiple params with regex",
			path:     "/users/{userID:[0-9]+}/{userName:[a-z]+}",
			expected: []string{"userID", "userName"},
		},
		{
			name:     "multiple params in the same section",
			path:     "/users/{userID}-{userName}",
			expected: []string{"userID", "userName"},
		},
		{
			name:     "multiple params in the same section with regex",
			path:     "/users/{userID:[0-9]+}-{userName:[a-z]+}",
			expected: []string{"userID", "userName"},
		},
		{
			name:     "mismatched brackets",
			path:     "/users/{userID}-{userName",
			wantErr:  true,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ExtractParamName(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractParamName(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}

			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("extractParamName(%q) = %v, want %v", tt.path, got, tt.expected)
			}
		})
	}
}

func TestSanitizePath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "empty path",
			path:     "",
			expected: "/",
		},
		{
			name:     "multiple double slashes",
			path:     "/users//{userID}//",
			expected: "/users/{userID}",
		},
		{
			name:     "no changes",
			path:     "/users/{userID}",
			expected: "/users/{userID}",
		},
		{
			name:     "trailing slash",
			path:     "/users/{userID}/",
			expected: "/users/{userID}",
		},
		{
			name:     "triple slash",
			path:     "/users///{userID}",
			expected: "/users/{userID}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := SanitizePath(tt.path)
			if got != tt.expected {
				t.Errorf("SanitizePath(%q) = %q, want %q", tt.path, got, tt.expected)
			}
		})
	}
}

func TestIsValidParameterName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		param string
		want  bool
	}{
		// Valid cases
		{
			name:  "single letter",
			param: "a",
			want:  true,
		},
		{
			name:  "multiple letters",
			param: "abc",
			want:  true,
		},
		{
			name:  "multiple underscores",
			param: "abc_def",
			want:  true,
		},
		{
			name:  "letters and digits",
			param: "abc123",
			want:  true,
		},
		{
			name:  "letters, digits, and underscores",
			param: "abc_123_def",
			want:  true,
		},
		// Invalid cases
		{
			name:  "empty",
			param: "",
			want:  false,
		},
		{
			name:  "single digit",
			param: "1",
			want:  false,
		},
		{
			name:  "single underscore",
			param: "_",
			want:  false,
		},
		{
			name:  "single dash",
			param: "-",
			want:  false,
		},
		{
			name:  "single dot",
			param: ".",
			want:  false,
		},
		{
			name:  "single slash",
			param: "/",
			want:  false,
		},
		{
			name:  "single backslash",
			param: "\\",
			want:  false,
		},
		{
			name:  "single pipe",
			param: "|",
			want:  false,
		},
		{
			name:  "single quote",
			param: "'",
			want:  false,
		},
		{
			name:  "double quote",
			param: "\"",
			want:  false,
		},
		{
			name:  "multiple digits",
			param: "123",
			want:  false,
		},
		{
			name:  "contains dash",
			param: "abc-def",
			want:  false,
		},
		{
			name:  "contains dot",
			param: "abc.def",
			want:  false,
		},
		{
			name:  "contains slash",
			param: "abc/def",
			want:  false,
		},
		{
			name:  "contains backslash",
			param: "abc\\def",
			want:  false,
		},
		{
			name:  "contains pipe",
			param: "abc|def",
			want:  false,
		},
		{
			name:  "single quotes in middle",
			param: "abc'def",
			want:  false,
		},
		{
			name:  "double quotes in middle",
			param: "abc\"def",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := IsValidParameterName(tt.param)
			if result != tt.want {
				t.Errorf("IsValidParameterName(%q) = %v, want %v", tt.param, result, tt.want)
			}
		})
	}
}
