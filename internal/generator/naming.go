package generator

import (
	"path/filepath"
	"strings"

	"github.com/oter/dotprompt-gen-go/internal/naming"
)

// FilenameToStructNames converts a filename to Go struct names.
func FilenameToStructNames(filename string) (string, string) {
	base := strings.TrimSuffix(filepath.Base(filename), ".prompt")

	// Convert snake_case to PascalCase
	pascal := naming.SnakeToPascalCase(base)

	return pascal + "Input", pascal + "Output"
}
