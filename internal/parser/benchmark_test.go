package parser

import (
	"testing"
)

// BenchmarkParsePicoschema benchmarks Picoschema parsing performance
func BenchmarkParsePicoschema(b *testing.B) {
	picoSchema := map[string]any{
		"name":        "string, the user name",
		"age":         "integer, user age",
		"email":       "string, user email address",
		"active":      "boolean, whether user is active",
		"tags":        "string(array), user tags",
		"priority":    "string(enum): [low, medium, high], task priority",
		"metadata":    "any, additional metadata",
		"score":       "number, user score",
		"created_at":  "string, creation timestamp",
		"updated_at?": "string, optional update timestamp",
	}

	requiredFields := []string{"name", "email", "active"}

	for b.Loop() {
		_, _, err := parsePicoschema(picoSchema, requiredFields, SchemaTypeOutput)
		if err != nil {
			b.Fatalf("Failed to parse schema: %v", err)
		}
	}
}

// BenchmarkParseJSONSchema benchmarks JSON Schema parsing performance
func BenchmarkParseJSONSchema(b *testing.B) {
	jsonSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{
				"type":        "string",
				"description": "User name",
			},
			"age": map[string]any{
				"type":        "integer",
				"description": "User age",
			},
			"email": map[string]any{
				"type":        "string",
				"description": "User email",
			},
			"active": map[string]any{
				"type":        "boolean",
				"description": "User active status",
			},
			"priority": map[string]any{
				"type":        "string",
				"enum":        []any{"low", "medium", "high"},
				"description": "Task priority",
			},
			"tags": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "string",
				},
				"description": "User tags",
			},
		},
		"required": []any{"name", "email", "active"},
	}

	requiredFields := []string{"name", "email", "active"}

	for b.Loop() {
		_, _, _, err := parseJSONSchemaWithStructs(jsonSchema, requiredFields, SchemaTypeOutput)
		if err != nil {
			b.Fatalf("Failed to parse schema: %v", err)
		}
	}
}

// BenchmarkDeeplyNestedJSONSchema benchmarks deeply nested JSON Schema parsing
func BenchmarkDeeplyNestedJSONSchema(b *testing.B) {
	deepSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"level1": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"level2": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"level3": map[string]any{
								"type": "object",
								"properties": map[string]any{
									"level4": map[string]any{
										"type": "object",
										"properties": map[string]any{
											"final_value": map[string]any{
												"type": "string",
											},
											"final_enum": map[string]any{
												"type": "string",
												"enum": []any{"option1", "option2", "option3"},
											},
										},
										"required": []any{"final_value", "final_enum"},
									},
								},
								"required": []any{"level4"},
							},
						},
						"required": []any{"level3"},
					},
				},
				"required": []any{"level2"},
			},
		},
		"required": []any{"level1"},
	}

	requiredFields := []string{"level1"}

	for b.Loop() {
		_, _, _, err := parseJSONSchemaWithStructs(deepSchema, requiredFields, SchemaTypeOutput)
		if err != nil {
			b.Fatalf("Failed to parse deeply nested schema: %v", err)
		}
	}
}

// BenchmarkSchemaFormatDetection benchmarks schema format detection
func BenchmarkSchemaFormatDetection(b *testing.B) {
	picoSchema := map[string]any{
		"name": "string, the user name",
		"age":  "integer, user age",
	}

	jsonSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{"type": "string"},
		},
	}

	b.Run("Picoschema", func(b *testing.B) {
		for b.Loop() {
			IsPicoschema(picoSchema)
		}
	})

	b.Run("JSONSchema", func(b *testing.B) {
		for b.Loop() {
			IsJSONSchema(jsonSchema)
		}
	})
}

// BenchmarkTypeConversion benchmarks type conversion functions
func BenchmarkTypeConversion(b *testing.B) {
	b.Run("PicoschemaTyeConversion", func(b *testing.B) {
		types := []string{"string", "integer", "number", "boolean", "any", "null"}
		for b.Loop() {
			for _, t := range types {
				convertPicoschemaTypeToGo(t)
			}
		}
	})

	b.Run("JSONSchemaTypeConversion", func(b *testing.B) {
		types := []string{"string", "integer", "number", "boolean", "array"}
		for b.Loop() {
			for _, t := range types {
				convertJSONSchemaTypeToGo(t)
			}
		}
	})
}
