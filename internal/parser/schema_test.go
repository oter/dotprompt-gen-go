package parser

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCodegenDeeplyNestedStructs tests parsing of schemas with 4 levels of nesting
func TestCodegenDeeplyNestedStructs(t *testing.T) {
	// Create a test schema with 4 levels of nesting
	deeplyNestedSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"level1": map[string]any{
				"type":        "object",
				"description": "First level object",
				"properties": map[string]any{
					"level2": map[string]any{
						"type":        "object",
						"description": "Second level object",
						"properties": map[string]any{
							"level3": map[string]any{
								"type":        "object",
								"description": "Third level object",
								"properties": map[string]any{
									"level4": map[string]any{
										"type":        "object",
										"description": "Fourth level object",
										"properties": map[string]any{
											"final_value": map[string]any{
												"type":        "string",
												"description": "Final nested value",
											},
											"final_enum": map[string]any{
												"type":        "string",
												"description": "Final nested enum",
												"enum":        []any{"option1", "option2", "option3"},
											},
										},
										"required": []any{"final_value", "final_enum"},
									},
									"level3_direct": map[string]any{
										"type":        "string",
										"description": "Direct value at level 3",
									},
								},
								"required": []any{"level4", "level3_direct"},
							},
							"level2_direct": map[string]any{
								"type":        "boolean",
								"description": "Direct value at level 2",
							},
						},
						"required": []any{"level3", "level2_direct"},
					},
					"level1_direct": map[string]any{
						"type":        "integer",
						"description": "Direct value at level 1",
					},
				},
				"required": []any{"level2", "level1_direct"},
			},
			"root_value": map[string]any{
				"type":        "string",
				"description": "Root level value",
			},
		},
		"required": []any{"level1", "root_value"},
	}

	fields, enums, structs, err := ParseSchemaWithStructs(deeplyNestedSchema, []string{"level1", "root_value"}, SchemaTypeOutput)
	require.NoError(t, err, "Failed to parse deeply nested schema")

	// Verify we have the expected number of fields at root level
	assert.Len(t, fields, 2, "Expected 2 root fields")

	// Verify we have all nested structs (Level1, Level1Level2, Level1Level2Level3, Level1Level2Level3Level4)
	expectedStructCount := 4
	if !assert.Len(t, structs, expectedStructCount, "Expected %d nested structs", expectedStructCount) {
		for i, s := range structs {
			t.Logf("Struct %d: %s", i, s.Name)
		}
	}

	// Verify we have the final enum
	expectedEnumCount := 1
	if !assert.Len(t, enums, expectedEnumCount, "Expected %d enums", expectedEnumCount) {
		for i, e := range enums {
			t.Logf("Enum %d: %s", i, e.Name)
		}
	}

	// Verify struct names are unique and correctly prefixed
	structNames := make(map[string]bool)
	for _, s := range structs {
		assert.False(t, structNames[s.Name], "Duplicate struct name: %s", s.Name)
		structNames[s.Name] = true
	}

	// Verify the deepest struct has the final fields
	var deepestStruct *struct {
		Name   string
		Fields []struct {
			Name   string
			GoType string
		}
	}
	for i := range structs {
		if strings.Contains(structs[i].Name, "Level4") {
			deepestStruct = &struct {
				Name   string
				Fields []struct {
					Name   string
					GoType string
				}
			}{
				Name: structs[i].Name,
				Fields: make([]struct {
					Name   string
					GoType string
				}, len(structs[i].Fields)),
			}
			for j, field := range structs[i].Fields {
				deepestStruct.Fields[j].Name = field.Name
				deepestStruct.Fields[j].GoType = field.GoType
			}
			break
		}
	}

	require.NotNil(t, deepestStruct, "Could not find the deepest nested struct (Level4)")

	// Check that the deepest struct has the expected fields
	expectedDeepFields := map[string]string{
		"FinalValue": "string",
		"FinalEnum":  "FinalEnumEnum",
	}

	assert.Len(t, deepestStruct.Fields, len(expectedDeepFields), "Expected specific number of fields in deepest struct")

	for _, field := range deepestStruct.Fields {
		expectedType, exists := expectedDeepFields[field.Name]
		assert.True(t, exists, "Unexpected field in deepest struct: %s", field.Name)
		if exists {
			assert.Equal(t, expectedType, field.GoType, "Field %s type mismatch", field.Name)
		}
	}

	t.Logf("Successfully parsed %d levels of nesting with %d structs and %d enums", 4, len(structs), len(enums))
}

// TestCodegenRealisticDeeplyNestedStructs tests realistic nested schema parsing
func TestCodegenRealisticDeeplyNestedStructs(t *testing.T) {
	// Create a realistic schema similar to our habit classification with deeper nesting
	realisticSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"classification": map[string]any{
				"type":        "object",
				"description": "Habit classification results",
				"properties": map[string]any{
					"evaluation": map[string]any{
						"type":        "object",
						"description": "Habit evaluation",
						"properties": map[string]any{
							"validity": map[string]any{
								"type":        "object",
								"description": "Habit assessments",
								"properties": map[string]any{
									"checks": map[string]any{
										"type":        "object",
										"description": "Individual habit assessments",
										"properties": map[string]any{
											"is_valid": map[string]any{
												"type":        "boolean",
												"description": "Is valid habit",
											},
											"reason_code": map[string]any{
												"type":        "string",
												"enum":        []any{"healthy", "needs_improvement", "harmful", "unclear"},
												"description": "Assessment code for habit evaluation",
											},
										},
										"required": []any{"is_valid", "reason_code"},
									},
									"confidence": map[string]any{
										"type":        "number",
										"description": "Confidence score",
									},
								},
								"required": []any{"checks", "confidence"},
							},
							"explanation": map[string]any{
								"type":        "string",
								"description": "Explanation of evaluation",
							},
						},
						"required": []any{"validity", "explanation"},
					},
					"categories": map[string]any{
						"type":        "object",
						"description": "Category classification",
						"properties": map[string]any{
							"primary": map[string]any{
								"type":        "string",
								"enum":        []any{"physical", "mental", "social", "creative"},
								"description": "Primary category",
							},
							"secondary": map[string]any{
								"type":        "string",
								"enum":        []any{"daily", "weekly", "monthly", "occasional"},
								"description": "Secondary category",
							},
						},
						"required": []any{"primary"},
					},
				},
				"required": []any{"evaluation", "categories"},
			},
		},
		"required": []any{"classification"},
	}

	_, enums, structs, err := ParseSchemaWithStructs(realisticSchema, []string{"classification"}, SchemaTypeOutput)
	require.NoError(t, err, "Failed to parse realistic nested schema")

	// Verify we have at least 4 levels of nesting: Classification -> Evaluation -> Validity -> Checks
	expectedMinStructs := 4
	if !assert.GreaterOrEqual(t, len(structs), expectedMinStructs, "Expected at least %d nested structs", expectedMinStructs) {
		for i, s := range structs {
			t.Logf("Struct %d: %s with %d fields", i, s.Name, len(s.Fields))
		}
	}

	// Verify we have the expected enums (ReasonCodeEnum, PrimaryEnum, SecondaryEnum)
	expectedMinEnums := 3
	if !assert.GreaterOrEqual(t, len(enums), expectedMinEnums, "Expected at least %d enums", expectedMinEnums) {
		for i, e := range enums {
			t.Logf("Enum %d: %s with %d values", i, e.Name, len(e.Values))
		}
	}

	t.Logf("Successfully parsed realistic nested structure with %d structs and %d enums", len(structs), len(enums))
}

// TestSchemaFormatDetection tests detection of different schema formats
func TestSchemaFormatDetection(t *testing.T) {
	// Test Picoschema detection
	picoSchema := map[string]any{
		"name": "string, the user name",
		"age":  "integer, user age",
	}

	assert.True(t, IsPicoschema(picoSchema), "Failed to detect Picoschema format")

	// Test JSON Schema detection
	jsonSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{
				"type": "string",
			},
		},
	}

	assert.True(t, IsJSONSchema(jsonSchema), "Failed to detect JSON Schema format")

	assert.False(t, IsPicoschema(jsonSchema), "Incorrectly detected JSON Schema as Picoschema")
}

// TestFieldOrderingConsistency tests that fields are generated in consistent order
func TestFieldOrderingConsistency(t *testing.T) {
	// Test Picoschema field ordering
	picoschema := map[string]any{
		"zebra_field":  "string, last field alphabetically",
		"alpha_field":  "string, first field alphabetically",
		"middle_field": "string, middle field alphabetically",
		"beta_field":   "string, second field alphabetically",
	}

	// Generate fields multiple times to ensure consistent ordering
	var fieldOrders [][]string
	for range 5 {
		fields, _, _, err := ParseSchemaWithStructs(picoschema, []string{}, SchemaTypeInput)
		require.NoError(t, err, "Failed to parse picoschema")

		var fieldNames []string
		for _, field := range fields {
			fieldNames = append(fieldNames, field.Name)
		}
		fieldOrders = append(fieldOrders, fieldNames)
	}

	// All field orders should be identical
	expectedOrder := []string{"AlphaField", "BetaField", "MiddleField", "ZebraField"}
	for i, fieldOrder := range fieldOrders {
		assert.Len(t, fieldOrder, len(expectedOrder), "Run %d: expected %d fields", i+1, len(expectedOrder))
		for j, expectedName := range expectedOrder {
			if j < len(fieldOrder) {
				assert.Equal(t, expectedName, fieldOrder[j], "Run %d: expected field %d to be %s", i+1, j, expectedName)
			}
		}
	}

	// Test JSON Schema field ordering
	jsonSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"zebra_prop": map[string]any{
				"type":        "string",
				"description": "Last field alphabetically",
			},
			"alpha_prop": map[string]any{
				"type":        "string",
				"description": "First field alphabetically",
			},
			"middle_prop": map[string]any{
				"type":        "string",
				"description": "Middle field alphabetically",
			},
		},
	}

	// Generate fields multiple times to ensure consistent ordering
	var jsonFieldOrders [][]string
	for range 5 {
		fields, _, _, err := ParseSchemaWithStructs(jsonSchema, []string{}, SchemaTypeOutput)
		require.NoError(t, err, "Failed to parse JSON schema")

		var fieldNames []string
		for _, field := range fields {
			fieldNames = append(fieldNames, field.Name)
		}
		jsonFieldOrders = append(jsonFieldOrders, fieldNames)
	}

	// All field orders should be identical
	expectedJSONOrder := []string{"AlphaProp", "MiddleProp", "ZebraProp"}
	for i, fieldOrder := range jsonFieldOrders {
		assert.Len(t, fieldOrder, len(expectedJSONOrder), "JSON Run %d: expected %d fields", i+1, len(expectedJSONOrder))
		for j, expectedName := range expectedJSONOrder {
			if j < len(fieldOrder) {
				assert.Equal(t, expectedName, fieldOrder[j], "JSON Run %d: expected field %d to be %s", i+1, j, expectedName)
			}
		}
	}
}
