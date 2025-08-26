package main

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// ParsePromptFile parses a dotprompt file and returns a PromptFile
func ParsePromptFile(filepath string) (*PromptFile, error) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filepath, err)
	}

	return ParsePromptContent(string(content), filepath)
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
