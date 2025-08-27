package generator

import (
	"path/filepath"
	"strings"
)

// FilenameToStructNames converts a filename to Go struct names
func FilenameToStructNames(filename string) (request, response string) {
	base := strings.TrimSuffix(filepath.Base(filename), ".prompt")

	// Convert snake_case to PascalCase
	pascal := SnakeToPascalCase(base)

	return pascal + "Request", pascal + "Response"
}

// SnakeToPascalCase converts snake_case to PascalCase
func SnakeToPascalCase(s string) string {
	if s == "" {
		return s
	}

	parts := strings.Split(s, "_")
	var result strings.Builder

	for _, part := range parts {
		if len(part) > 0 {
			result.WriteString(strings.ToUpper(part[:1]))
			if len(part) > 1 {
				result.WriteString(part[1:])
			}
		}
	}

	return result.String()
}
