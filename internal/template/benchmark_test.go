package template

import (
	"testing"
)

// BenchmarkValidateHandlebarsTemplate benchmarks template validation performance
func BenchmarkValidateHandlebarsTemplate(b *testing.B) {
	templates := []string{
		"Hello {{name}}!",
		"{{role \"system\"}}You are helpful.{{role \"user\"}}{{question}}",
		"Items: {{#each items}}{{this}} {{/each}}",
		"User: {{user.name}} ({{user.email}}) - Status: {{status}}",
		"{{role \"system\"}}Process {{#each items}}{{name}}: {{value}} {{/each}}{{role \"user\"}}{{query}}",
	}

	for b.Loop() {
		for _, template := range templates {
			result := ValidateHandlebarsTemplate(template)
			if !result.Valid {
				b.Fatalf("Template validation failed: %v", result.Errors)
			}
		}
	}
}

// BenchmarkValidateComplexTemplate benchmarks validation of complex templates
func BenchmarkValidateComplexTemplate(b *testing.B) {
	complexTemplate := `{{role "system"}}
You are a wise habit classification sage who understands the deeper patterns of human transformation.

Classify the given habit into one of these domains:
{{#each categories}}
- **{{name}}**: {{description}}
{{/each}}

Consider the following factors:
{{#each factors}}
{{@index}}. {{this}}
{{/each}}

{{role "user"}}
Habit to classify: {{habit}}

Additional context:
- User: {{user.name}} ({{user.email}})
- Priority: {{priority}}
- Tags: {{#each tags}}{{this}}{{#unless @last}}, {{/unless}}{{/each}}
- Metadata: {{metadata}}`

	for b.Loop() {
		result := ValidateHandlebarsTemplate(complexTemplate)
		if !result.Valid {
			b.Fatalf("Complex template validation failed: %v", result.Errors)
		}
	}
}

// BenchmarkValidateVariablesAgainstSchema benchmarks schema validation
func BenchmarkValidateVariablesAgainstSchema(b *testing.B) {
	variables := []string{
		"name", "age", "email", "user.name", "user.email", "user.profile.bio",
		"settings.theme", "preferences.language", "metadata", "tags", "status",
	}

	schema := map[string]any{
		"properties": map[string]any{
			"name":  map[string]any{"type": "string"},
			"age":   map[string]any{"type": "integer"},
			"email": map[string]any{"type": "string"},
			"user": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name":    map[string]any{"type": "string"},
					"email":   map[string]any{"type": "string"},
					"profile": map[string]any{"type": "object"},
				},
			},
			"settings": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"theme": map[string]any{"type": "string"},
				},
			},
			"preferences": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"language": map[string]any{"type": "string"},
				},
			},
			"metadata": map[string]any{"type": "object"},
			"tags":     map[string]any{"type": "array"},
			"status":   map[string]any{"type": "string"},
		},
	}

	for b.Loop() {
		ValidateVariablesAgainstSchema(variables, schema)
	}
}

// BenchmarkValidateHelpers benchmarks helper validation
func BenchmarkValidateHelpers(b *testing.B) {
	helpers := []HelperUsage{
		{Name: "role", Parameters: []string{"system"}},
		{Name: "role", Parameters: []string{"user"}},
		{Name: "role", Parameters: []string{"assistant"}},
	}

	blockHelpers := []BlockHelperUsage{
		{Name: "each", Parameters: []string{"items"}},
		{Name: "if", Parameters: []string{"condition"}},
		{Name: "unless", Parameters: []string{"condition"}},
	}

	for b.Loop() {
		ValidateHelpers(helpers, blockHelpers)
	}
}
