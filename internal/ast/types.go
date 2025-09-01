package ast

import (
	"github.com/oter/dotprompt-gen-go/internal/template"
)

// PromptFile represents a parsed dotprompt file.
type PromptFile struct {
	Filename    string          // File identifier
	Frontmatter FrontmatterData `yaml:",inline"` // YAML metadata defining structure
	Template    string          // Template content
}

// FrontmatterData represents the YAML frontmatter in a dotprompt file.
type FrontmatterData struct {
	Model  string     `yaml:"model"`
	Input  SchemaSpec `yaml:"input"`
	Output SchemaSpec `yaml:"output"`
	Config any        `yaml:"config"`
}

// SchemaSpec represents input/output schema specification.
type SchemaSpec struct {
	Schema   any      `yaml:"schema"`
	Default  any      `yaml:"default"`
	Required []string `yaml:"required"`
}

// HasSchema checks if the prompt file has input or output schema.
func (pf *PromptFile) HasSchema() bool {
	return pf.Frontmatter.Input.Schema != nil || pf.Frontmatter.Output.Schema != nil
}

// GetInputSchema returns the input schema if available.
func (pf *PromptFile) GetInputSchema() any {
	return pf.Frontmatter.Input.Schema
}

// GetOutputSchema returns the output schema if available.
func (pf *PromptFile) GetOutputSchema() any {
	return pf.Frontmatter.Output.Schema
}

// GetRequiredInputFields returns required input fields.
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

// GetRequiredOutputFields returns required output fields.
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

// ValidateTemplate validates the handlebars template against the input schema.
func (pf *PromptFile) ValidateTemplate() *template.ValidationResult {
	return template.ValidateHandlebarsTemplate(pf.Template)
}

// ValidateTemplateWithSchema validates template variables against input schema.
func (pf *PromptFile) ValidateTemplateWithSchema() []template.ValidationError {
	var allErrors []template.ValidationError

	// First validate template syntax
	result := pf.ValidateTemplate()
	if !result.Valid {
		allErrors = append(allErrors, result.Errors...)

		return allErrors
	}

	// Validate variables against input schema if schema exists
	if inputSchema := pf.GetInputSchema(); inputSchema != nil {
		if schemaMap, ok := inputSchema.(map[string]any); ok {
			schemaErrors := template.ValidateVariablesAgainstSchema(result.Variables, schemaMap)
			allErrors = append(allErrors, schemaErrors...)
		}
	}

	// Validate helper functions
	helperErrors := template.ValidateHelpers(result.Helpers, result.BlockHelpers)
	allErrors = append(allErrors, helperErrors...)

	return allErrors
}

// GetTemplateVariables extracts all variables used in the template.
func (pf *PromptFile) GetTemplateVariables() []string {
	result := pf.ValidateTemplate()

	return result.Variables
}

// GetTemplateHelpers extracts all helpers used in the template.
func (pf *PromptFile) GetTemplateHelpers() []template.HelperUsage {
	result := pf.ValidateTemplate()

	return result.Helpers
}
