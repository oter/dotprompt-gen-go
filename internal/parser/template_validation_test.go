package parser

import (
	"testing"

	"github.com/oter/dotprompt-gen-go/internal/ast"
	"github.com/stretchr/testify/assert"
)

func TestValidateTemplate(t *testing.T) {
	tests := []struct {
		name      string
		template  string
		wantValid bool
		wantVars  []string
	}{
		{
			name:      "valid template with variables",
			template:  "Hello {{name}}, you are {{age}} years old!",
			wantValid: true,
			wantVars:  []string{"name", "age"},
		},
		{
			name:      "template with role helper",
			template:  "{{role \"system\"}}You are helpful.{{role \"user\"}}{{question}}",
			wantValid: true,
			wantVars:  []string{"question"},
		},
		{
			name:      "invalid template syntax",
			template:  "Hello {{name!",
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pf := &ast.PromptFile{
				Template: tt.template,
			}

			result := pf.ValidateTemplate()

			assert.Equal(t, tt.wantValid, result.Valid, "Expected valid=%v, got valid=%v, errors: %v",
				tt.wantValid, result.Valid, result.Errors)

			if tt.wantValid {
				assert.Len(t, result.Variables, len(tt.wantVars), "Expected %d variables, got %v", len(tt.wantVars), result.Variables)
			}
		})
	}
}

func TestValidateTemplateWithSchema(t *testing.T) {
	tests := []struct {
		name       string
		template   string
		schema     map[string]any
		wantErrors int
	}{
		{
			name:     "valid template with matching schema",
			template: "Hello {{name}}, you are {{age}} years old!",
			schema: map[string]any{
				"properties": map[string]any{
					"name": map[string]any{"type": "string"},
					"age":  map[string]any{"type": "integer"},
				},
			},
			wantErrors: 0,
		},
		{
			name:     "template with undefined variable",
			template: "Hello {{name}}, you are {{undefined_var}} years old!",
			schema: map[string]any{
				"properties": map[string]any{
					"name": map[string]any{"type": "string"},
				},
			},
			wantErrors: 1, // undefined_var not in schema
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pf := &ast.PromptFile{
				Frontmatter: ast.FrontmatterData{
					Input: ast.SchemaSpec{
						Schema: tt.schema,
					},
				},
				Template: tt.template,
			}

			errors := pf.ValidateTemplateWithSchema()

			assert.Len(t, errors, tt.wantErrors, "Expected %d errors, got %v", tt.wantErrors, errors)
		})
	}
}

func TestGetTemplateVariables(t *testing.T) {
	tests := []struct {
		name     string
		template string
		want     []string
	}{
		{
			name:     "simple variables",
			template: "Hello {{name}}, age {{age}}",
			want:     []string{"name", "age"},
		},
		{
			name:     "nested variables",
			template: "User: {{user.name}} ({{user.email}})",
			want:     []string{"user.name", "user.email"},
		},
		{
			name:     "variables in block helpers",
			template: "{{#each items}}{{name}} - {{value}}{{/each}}",
			want:     []string{"items", "name", "value"},
		},
		{
			name:     "no variables",
			template: "Static content only",
			want:     []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pf := &ast.PromptFile{
				Template: tt.template,
			}

			got := pf.GetTemplateVariables()

			if !assert.Len(t, got, len(tt.want), "Expected %d variables, got %v", len(tt.want), got) {
				return
			}

			for _, wantVar := range tt.want {
				assert.Contains(t, got, wantVar, "Expected variable '%s' not found in %v", wantVar, got)
			}
		})
	}
}

func TestGetTemplateHelpers(t *testing.T) {
	tests := []struct {
		name        string
		template    string
		wantHelpers []string
	}{
		{
			name:        "role helpers",
			template:    "{{role \"system\"}}Hello{{role \"user\"}}Question",
			wantHelpers: []string{"role", "role"},
		},
		{
			name:        "no helpers",
			template:    "Hello {{name}}",
			wantHelpers: []string{},
		},
		{
			name:        "mixed helpers and variables",
			template:    "{{role \"system\"}}Process {{data}}",
			wantHelpers: []string{"role"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pf := &ast.PromptFile{
				Template: tt.template,
			}

			helpers := pf.GetTemplateHelpers()

			if !assert.Len(t, helpers, len(tt.wantHelpers), "Expected %d helpers", len(tt.wantHelpers)) {
				return
			}

			for i, wantHelper := range tt.wantHelpers {
				if i < len(helpers) {
					assert.Equal(t, wantHelper, helpers[i].Name, "Helper[%d] name mismatch", i)
				}
			}
		})
	}
}
