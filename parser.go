package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ParsePromptFile parses a dotprompt file and returns a PromptFile
func ParsePromptFile(filePath string) (*PromptFile, error) {
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

// ParsePromptContent parses dotprompt content and returns a PromptFile
func ParsePromptContent(content, filename string) (*PromptFile, error) {
	// Split by frontmatter delimiters
	parts := strings.Split(content, "---")
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid dotprompt format: missing frontmatter delimiters")
	}

	// Parse YAML frontmatter
	frontmatterContent := strings.TrimSpace(parts[1])
	var frontmatter FrontmatterData
	if err := yaml.Unmarshal([]byte(frontmatterContent), &frontmatter); err != nil {
		return nil, fmt.Errorf("failed to parse YAML frontmatter: %w", err)
	}

	// Extract template content (everything after the second ---)
	templateParts := parts[2:]
	template := strings.TrimSpace(strings.Join(templateParts, "---"))

	return &PromptFile{
		Frontmatter: frontmatter,
		Template:    template,
		Filename:    filename,
	}, nil
}

// HasSchema checks if the prompt file has input or output schema
func (pf *PromptFile) HasSchema() bool {
	return pf.Frontmatter.Input.Schema != nil || pf.Frontmatter.Output.Schema != nil
}

// GetInputSchema returns the input schema if available
func (pf *PromptFile) GetInputSchema() any {
	return pf.Frontmatter.Input.Schema
}

// GetOutputSchema returns the output schema if available
func (pf *PromptFile) GetOutputSchema() any {
	return pf.Frontmatter.Output.Schema
}

// GetRequiredInputFields returns required input fields
func (pf *PromptFile) GetRequiredInputFields() []string {
	// Check if schema has required fields
	if schema, ok := pf.Frontmatter.Input.Schema.(map[string]any); ok {
		if required, ok := schema["required"].([]any); ok {
			var fields []string
			for _, field := range required {
				if fieldStr, ok := field.(string); ok {
					fields = append(fields, fieldStr)
				}
			}
			return fields
		}
	}

	// Fallback to frontmatter required fields
	return pf.Frontmatter.Input.Required
}

// GetRequiredOutputFields returns required output fields
func (pf *PromptFile) GetRequiredOutputFields() []string {
	// Check if schema has required fields
	if schema, ok := pf.Frontmatter.Output.Schema.(map[string]any); ok {
		if required, ok := schema["required"].([]any); ok {
			var fields []string
			for _, field := range required {
				if fieldStr, ok := field.(string); ok {
					fields = append(fields, fieldStr)
				}
			}
			return fields
		}
	}

	// Fallback to frontmatter required fields
	return pf.Frontmatter.Output.Required
}
