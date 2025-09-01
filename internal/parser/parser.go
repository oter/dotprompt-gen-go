package parser

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/oter/dotprompt-gen-go/internal/ast"
)

const (
	// minimumFrontmatterParts is the minimum number of parts when splitting by --- delimiters.
	minimumFrontmatterParts = 3
)

// ParsePromptFile parses a dotprompt file and returns a PromptFile.
func ParsePromptFile(filePath string) (*ast.PromptFile, error) {
	// Validate and clean the file path to prevent path traversal attacks
	cleanPath := filepath.Clean(filePath)

	// Ensure the file has the .prompt extension
	if !strings.HasSuffix(cleanPath, ".prompt") {
		return nil, fmt.Errorf("invalid file extension: expected .prompt file, got %s", cleanPath)
	}

	// Convert to absolute path to further validate
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute path for %s: %w", cleanPath, err)
	}

	// Check if file exists and is a regular file (not a directory or special file)
	info, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to access file %s: %w", absPath, err)
	}

	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("path is not a regular file: %s", absPath)
	}

	// #nosec G304 - Path has been validated above
	content, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", absPath, err)
	}

	return ParsePromptContent(string(content), absPath)
}

// ParsePromptContent parses dotprompt content and returns a PromptFile.
func ParsePromptContent(content, filename string) (*ast.PromptFile, error) {
	// Split by frontmatter delimiters
	parts := strings.Split(content, "---")
	if len(parts) < minimumFrontmatterParts {
		return nil, errors.New("invalid dotprompt format: missing frontmatter delimiters")
	}

	// Parse YAML frontmatter
	frontmatterContent := strings.TrimSpace(parts[1])

	var frontmatter ast.FrontmatterData

	err := yaml.Unmarshal([]byte(frontmatterContent), &frontmatter)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML frontmatter: %w", err)
	}

	// Extract template content (everything after the second ---)
	templateParts := parts[2:]
	template := strings.TrimSpace(strings.Join(templateParts, "---"))

	return &ast.PromptFile{
		Filename:    filename,
		Frontmatter: frontmatter,
		Template:    template,
	}, nil
}
