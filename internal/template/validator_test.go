package template

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateHandlebarsTemplate_ValidSyntax(t *testing.T) {
	tests := []struct {
		name     string
		template string
		wantVars []string
	}{
		{
			name:     "simple variable",
			template: "Hello {{name}}!",
			wantVars: []string{"name"},
		},
		{
			name:     "multiple variables",
			template: "{{name}} is {{age}} years old",
			wantVars: []string{"name", "age"},
		},
		{
			name:     "nested variable",
			template: "Email: {{user.email}}",
			wantVars: []string{"user.email"},
		},
		{
			name:     "no variables",
			template: "Static content only",
			wantVars: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateHandlebarsTemplate(tt.template)

			assert.True(t, result.Valid, "Expected valid template, got errors: %v", result.Errors)
			assert.Len(t, result.Variables, len(tt.wantVars), "Expected %d variables", len(tt.wantVars))

			for i, want := range tt.wantVars {
				if i < len(result.Variables) {
					assert.Equal(t, want, result.Variables[i], "Expected variable[%d] = %s", i, want)
				}
			}
		})
	}
}

func TestValidateHandlebarsTemplate_InvalidSyntax(t *testing.T) {
	tests := []struct {
		name     string
		template string
	}{
		{
			name:     "unmatched opening brace",
			template: "Hello {{name!",
		},
		{
			name:     "malformed block helper",
			template: "{{#each items}}content{{/different}}",
		},
		{
			name:     "malformed expression with invalid helper",
			template: "Hello {{}}!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateHandlebarsTemplate(tt.template)

			assert.False(t, result.Valid, "Expected invalid template, but validation passed")
			assert.NotEmpty(t, result.Errors, "Expected syntax errors, but got none")

			if len(result.Errors) > 0 {
				assert.Equal(t, "syntax", result.Errors[0].Type, "Expected syntax error type")
			}
		})
	}
}

func TestValidateHandlebarsTemplate_RoleHelper(t *testing.T) {
	tests := []struct {
		name        string
		template    string
		wantHelpers []HelperUsage
		wantValid   bool
	}{
		{
			name:     "valid system role",
			template: "{{role \"system\"}}You are a helpful assistant.",
			wantHelpers: []HelperUsage{
				{Name: "role", Parameters: []string{"system"}},
			},
			wantValid: true,
		},
		{
			name:     "valid user role",
			template: "{{role \"user\"}}What is AI?",
			wantHelpers: []HelperUsage{
				{Name: "role", Parameters: []string{"user"}},
			},
			wantValid: true,
		},
		{
			name:     "multiple role helpers",
			template: "{{role \"system\"}}System prompt.{{role \"user\"}}User query.",
			wantHelpers: []HelperUsage{
				{Name: "role", Parameters: []string{"system"}},
				{Name: "role", Parameters: []string{"user"}},
			},
			wantValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateHandlebarsTemplate(tt.template)

			assert.Equal(t, tt.wantValid, result.Valid, "Expected valid=%v, got valid=%v, errors: %v",
				tt.wantValid, result.Valid, result.Errors)

			if !assert.Len(t, result.Helpers, len(tt.wantHelpers), "Expected %d helpers", len(tt.wantHelpers)) {
				return
			}

			for i, want := range tt.wantHelpers {
				got := result.Helpers[i]
				assert.Equal(t, want.Name, got.Name, "Helper[%d] name mismatch", i)
				if !assert.Len(t, got.Parameters, len(want.Parameters), "Helper[%d] params length mismatch", i) {
					continue
				}
				for j, wantParam := range want.Parameters {
					assert.Equal(t, wantParam, got.Parameters[j], "Helper[%d] param[%d] mismatch", i, j)
				}
			}
		})
	}
}

func TestValidateHandlebarsTemplate_EachHelper(t *testing.T) {
	tests := []struct {
		name             string
		template         string
		wantBlockHelpers []BlockHelperUsage
		wantVars         []string
	}{
		{
			name:     "each with array",
			template: "{{#each tags}}{{this}} {{/each}}",
			wantBlockHelpers: []BlockHelperUsage{
				{Name: "each", Parameters: []string{"tags"}},
			},
			wantVars: []string{"tags", "this"},
		},
		{
			name:     "each with nested content",
			template: "{{#each items}}Item: {{name}} - {{value}}{{/each}}",
			wantBlockHelpers: []BlockHelperUsage{
				{Name: "each", Parameters: []string{"items"}},
			},
			wantVars: []string{"items", "name", "value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateHandlebarsTemplate(tt.template)

			assert.True(t, result.Valid, "Expected valid template, got errors: %v", result.Errors)
			assert.Len(t, result.BlockHelpers, len(tt.wantBlockHelpers), "Expected %d block helpers", len(tt.wantBlockHelpers))

			for i, want := range tt.wantBlockHelpers {
				if i >= len(result.BlockHelpers) {
					break
				}
				got := result.BlockHelpers[i]
				assert.Equal(t, want.Name, got.Name, "BlockHelper[%d] name mismatch", i)
				if !assert.Len(t, got.Parameters, len(want.Parameters), "BlockHelper[%d] params length mismatch", i) {
					continue
				}
				for j, wantParam := range want.Parameters {
					assert.Equal(t, wantParam, got.Parameters[j], "BlockHelper[%d] param[%d] mismatch", i, j)
				}
			}

			// Check variables
			assert.Len(t, result.Variables, len(tt.wantVars), "Expected %d variables", len(tt.wantVars))
		})
	}
}

func TestValidateVariablesAgainstSchema(t *testing.T) {
	schema := map[string]any{
		"properties": map[string]any{
			"name": map[string]any{"type": "string"},
			"age":  map[string]any{"type": "integer"},
			"user": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"email": map[string]any{"type": "string"},
				},
			},
		},
	}

	tests := []struct {
		name          string
		variables     []string
		wantErrorVars []string
	}{
		{
			name:          "all valid variables",
			variables:     []string{"name", "age", "user.email"},
			wantErrorVars: []string{},
		},
		{
			name:          "undefined variable",
			variables:     []string{"name", "undefined_var"},
			wantErrorVars: []string{"undefined_var"},
		},
		{
			name:          "special variables ignored",
			variables:     []string{"name", "this", "@index"},
			wantErrorVars: []string{},
		},
		{
			name:          "nested variable with valid root",
			variables:     []string{"user.email", "user.name"},
			wantErrorVars: []string{},
		},
		{
			name:          "multiple undefined variables",
			variables:     []string{"name", "bad1", "bad2"},
			wantErrorVars: []string{"bad1", "bad2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateVariablesAgainstSchema(tt.variables, schema)

			assert.Len(t, errors, len(tt.wantErrorVars), "Expected %d errors", len(tt.wantErrorVars))

			errorVars := make(map[string]bool)
			for _, err := range errors {
				// Extract variable name from error message
				if strings.Contains(err.Message, "Variable '") {
					start := strings.Index(err.Message, "'") + 1
					end := strings.Index(err.Message[start:], "'")
					if end > 0 {
						varName := err.Message[start : start+end]
						errorVars[varName] = true
					}
				}
			}

			for _, wantVar := range tt.wantErrorVars {
				assert.True(t, errorVars[wantVar], "Expected error for variable '%s', but didn't find it", wantVar)
			}
		})
	}
}

func TestValidateHelpers_RoleValidation(t *testing.T) {
	tests := []struct {
		name          string
		helpers       []HelperUsage
		wantErrors    int
		wantErrorMsgs []string
	}{
		{
			name: "valid role",
			helpers: []HelperUsage{
				{Name: "role", Parameters: []string{"system"}},
			},
			wantErrors: 0,
		},
		{
			name: "invalid role",
			helpers: []HelperUsage{
				{Name: "role", Parameters: []string{"invalid"}},
			},
			wantErrors:    1,
			wantErrorMsgs: []string{"Invalid role 'invalid'"},
		},
		{
			name: "role with no parameters",
			helpers: []HelperUsage{
				{Name: "role", Parameters: []string{}},
			},
			wantErrors:    1,
			wantErrorMsgs: []string{"role helper expects 1 parameter"},
		},
		{
			name: "role with too many parameters",
			helpers: []HelperUsage{
				{Name: "role", Parameters: []string{"system", "extra"}},
			},
			wantErrors:    1,
			wantErrorMsgs: []string{"role helper expects 1 parameter"},
		},
		{
			name: "unknown helper",
			helpers: []HelperUsage{
				{Name: "unknown", Parameters: []string{"param"}},
			},
			wantErrors:    1,
			wantErrorMsgs: []string{"Unknown helper function 'unknown'"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateHelpers(tt.helpers, []BlockHelperUsage{})

			assert.Len(t, errors, tt.wantErrors, "Expected %d errors", tt.wantErrors)

			for i, wantMsg := range tt.wantErrorMsgs {
				if !assert.Greater(t, len(errors), i, "Expected error message containing '%s', but only got %d errors", wantMsg, len(errors)) {
					continue
				}
				assert.Contains(t, errors[i].Message, wantMsg, "Expected error message to contain '%s'", wantMsg)
			}
		})
	}
}

func TestValidateHelpers_EachValidation(t *testing.T) {
	tests := []struct {
		name         string
		blockHelpers []BlockHelperUsage
		wantErrors   int
	}{
		{
			name: "valid each helper",
			blockHelpers: []BlockHelperUsage{
				{Name: "each", Parameters: []string{"items"}},
			},
			wantErrors: 0,
		},
		{
			name: "each without parameters",
			blockHelpers: []BlockHelperUsage{
				{Name: "each", Parameters: []string{}},
			},
			wantErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateHelpers([]HelperUsage{}, tt.blockHelpers)

			assert.Len(t, errors, tt.wantErrors, "Expected %d errors", tt.wantErrors)
		})
	}
}

// Integration test with real prompt file templates
func TestValidateHandlebarsTemplate_RealExamples(t *testing.T) {
	tests := []struct {
		name      string
		template  string
		schema    map[string]any
		wantValid bool
	}{
		{
			name: "classify habits template",
			template: `{{role "system"}}
You are a wise habit classification sage.
{{role "user"}}
Habit to classify:
{{habit}}`,
			schema: map[string]any{
				"properties": map[string]any{
					"habit": map[string]any{"type": "string"},
				},
			},
			wantValid: true,
		},
		{
			name:     "array processing template",
			template: `Process tags: {{#each tags}}{{this}} {{/each}}`,
			schema: map[string]any{
				"properties": map[string]any{
					"tags": map[string]any{
						"type":  "array",
						"items": map[string]any{"type": "string"},
					},
				},
			},
			wantValid: true,
		},
		{
			name:     "simple variable template",
			template: `Process this data: {{name}} is {{age}} years old.`,
			schema: map[string]any{
				"properties": map[string]any{
					"name": map[string]any{"type": "string"},
					"age":  map[string]any{"type": "integer"},
				},
			},
			wantValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// First validate template syntax
			result := ValidateHandlebarsTemplate(tt.template)

			if tt.wantValid && !assert.True(t, result.Valid, "Template failed syntax validation: %v", result.Errors) {
				return
			}

			// Then validate against schema
			schemaErrors := ValidateVariablesAgainstSchema(result.Variables, tt.schema)
			helperErrors := ValidateHelpers(result.Helpers, result.BlockHelpers)

			allErrors := append(schemaErrors, helperErrors...)
			hasErrors := len(allErrors) > 0

			if tt.wantValid {
				assert.False(t, hasErrors, "Template validation failed: %v", allErrors)
			} else {
				assert.True(t, hasErrors, "Expected template validation to fail, but it passed")
			}
		})
	}
}
