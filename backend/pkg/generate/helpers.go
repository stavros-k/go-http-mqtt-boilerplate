package generate

import (
	"errors"
	"strings"
	"unicode"
)

// SanitizePath removes double slashes and trailing slashes from a path.
func SanitizePath(path string) string {
	cleanPath := path
	for strings.Contains(cleanPath, "//") {
		cleanPath = strings.ReplaceAll(cleanPath, "//", "/")
	}

	cleanPath = strings.TrimSuffix(cleanPath, "/")
	if cleanPath == "" {
		cleanPath = "/"
	}

	return cleanPath
}

// ExtractParamName extracts parameter names from a path.
// It returns an error if the number of '{' and '}' braces is mismatched.
func ExtractParamName(path string) ([]string, error) {
	dirtyParams := []string{}
	cleanParams := []string{}

	openBracket := strings.Count(path, "{")

	closeBracket := strings.Count(path, "}")
	if openBracket != closeBracket {
		return nil, errors.New("mismatched number of '{' and '}' in path")
	}
	// Find the content between '{' and '}'
	// Examples:
	// - {userID} -> userID
	// - {userID:[0-9]+} -> userID:[0-9]+
	start := -1
	for i, ch := range path {
		if ch == '{' {
			start = i + 1
		} else if ch == '}' && start >= 0 {
			dirtyParams = append(dirtyParams, path[start:i])
			start = -1
		}
	}

	// Now split on ':' to remove any regex matchers
	// Examples:
	// - userID -> userID
	// - userID:[0-9]+ -> userID
	for _, param := range dirtyParams {
		parts := strings.Split(param, ":")

		param = parts[0]
		if param != "" {
			cleanParams = append(cleanParams, param)
		}
	}

	return cleanParams, nil
}

// IsASCIILetterString checks if a string contains only ASCII letters (a-z, A-Z).
func IsASCIILetterString(s string) bool {
	if len(s) == 0 {
		return false
	}

	for _, r := range s {
		if !isASCIILetter(r) {
			return false
		}
	}

	return true
}

// isASCIILetter checks if a rune is an ASCII letter (a-z, A-Z).
func isASCIILetter(r rune) bool {
	if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
		return true
	}

	return false
}

// IsValidParameterName validates that a parameter name:
// - Starts with a letter (a-z, A-Z)
// - Contains only letters, digits, and underscores.
func IsValidParameterName(name string) bool {
	if name == "" {
		return false
	}

	for i, r := range name {
		if i == 0 {
			if !isASCIILetter(r) {
				return false
			}

			continue
		}

		// Subsequent characters must be letters, digits, or underscores
		if !isASCIILetter(r) && !unicode.IsDigit(r) && r != '_' {
			return false
		}
	}

	return true
}
