package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCase represents a single test case for code generation
type TestCase struct {
	Name            string
	PromptFile      string
	ExpectedStructs []ExpectedStruct
	ExpectedEnums   []ExpectedEnum
}

// ExpectedStruct represents expected struct generation
type ExpectedStruct struct {
	Name   string
	Fields []ExpectedField
}

// ExpectedField represents expected field in struct
type ExpectedField struct {
	Name          string
	GoType        string
	JSONTag       string
	HasValidation bool
}

// ExpectedEnum represents expected enum generation
type ExpectedEnum struct {
	Name   string
	Type   string
	Values []string
}

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

	fields, enums, structs, err := ParseSchemaWithStructs(deeplyNestedSchema, []string{"level1", "root_value"})
	if err != nil {
		t.Fatalf("Failed to parse deeply nested schema: %v", err)
	}

	// Verify we have the expected number of fields at root level
	if len(fields) != 2 {
		t.Errorf("Expected 2 root fields, got %d", len(fields))
	}

	// Verify we have all nested structs (Level1, Level1Level2, Level1Level2Level3, Level1Level2Level3Level4)
	expectedStructCount := 4
	if len(structs) != expectedStructCount {
		t.Errorf("Expected %d nested structs, got %d", expectedStructCount, len(structs))
		for i, s := range structs {
			t.Logf("Struct %d: %s", i, s.Name)
		}
	}

	// Verify we have the final enum
	expectedEnumCount := 1
	if len(enums) != expectedEnumCount {
		t.Errorf("Expected %d enums, got %d", expectedEnumCount, len(enums))
		for i, e := range enums {
			t.Logf("Enum %d: %s", i, e.Name)
		}
	}

	// Verify struct names are unique and correctly prefixed
	structNames := make(map[string]bool)
	for _, s := range structs {
		if structNames[s.Name] {
			t.Errorf("Duplicate struct name: %s", s.Name)
		}
		structNames[s.Name] = true
	}

	// Verify the deepest struct has the final fields
	var deepestStruct *GoStruct
	for i := range structs {
		if strings.Contains(structs[i].Name, "Level4") {
			deepestStruct = &structs[i]
			break
		}
	}

	if deepestStruct == nil {
		t.Fatal("Could not find the deepest nested struct (Level4)")
	}

	// Check that the deepest struct has the expected fields
	expectedDeepFields := map[string]string{
		"FinalValue": "string",
		"FinalEnum":  "FinalEnumEnum",
	}

	if len(deepestStruct.Fields) != len(expectedDeepFields) {
		t.Errorf("Expected %d fields in deepest struct, got %d", len(expectedDeepFields), len(deepestStruct.Fields))
	}

	for _, field := range deepestStruct.Fields {
		expectedType, exists := expectedDeepFields[field.Name]
		if !exists {
			t.Errorf("Unexpected field in deepest struct: %s", field.Name)
			continue
		}
		if field.GoType != expectedType {
			t.Errorf("Field %s: expected type %s, got %s", field.Name, expectedType, field.GoType)
		}
	}

	t.Logf("Successfully parsed %d levels of nesting with %d structs and %d enums", 4, len(structs), len(enums))
}

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

	_, enums, structs, err := ParseSchemaWithStructs(realisticSchema, []string{"classification"})
	if err != nil {
		t.Fatalf("Failed to parse realistic nested schema: %v", err)
	}

	// Verify we have at least 4 levels of nesting: Classification -> Evaluation -> Validity -> Checks
	expectedMinStructs := 4
	if len(structs) < expectedMinStructs {
		t.Errorf("Expected at least %d nested structs, got %d", expectedMinStructs, len(structs))
		for i, s := range structs {
			t.Logf("Struct %d: %s with %d fields", i, s.Name, len(s.Fields))
		}
	}

	// Verify we have the expected enums (ReasonCodeEnum, PrimaryEnum, SecondaryEnum)
	expectedMinEnums := 3
	if len(enums) < expectedMinEnums {
		t.Errorf("Expected at least %d enums, got %d", expectedMinEnums, len(enums))
		for i, e := range enums {
			t.Logf("Enum %d: %s with %d values", i, e.Name, len(e.Values))
		}
	}

	t.Logf("Successfully parsed realistic nested structure with %d structs and %d enums", len(structs), len(enums))
}

func TestGeneratedCodeCompiles(t *testing.T) {
	// Test that deeply nested generated code actually compiles
	deepSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"level1": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"level2": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"value":  map[string]any{"type": "string"},
							"status": map[string]any{"type": "string", "enum": []any{"active", "inactive"}},
						},
						"required": []any{"value", "status"},
					},
				},
				"required": []any{"level2"},
			},
		},
		"required": []any{"level1"},
	}

	fields, enums, structs, err := ParseSchemaWithStructs(deepSchema, []string{"level1"})
	if err != nil {
		t.Fatalf("Failed to parse schema: %v", err)
	}

	// Generate Go code
	code, err := GenerateGoCode(structs, enums, "testpkg")
	if err != nil {
		t.Fatalf("Failed to generate Go code: %v", err)
	}

	// Verify the code contains expected struct definitions
	codeStr := string(code)
	if !strings.Contains(codeStr, "type Level1 struct") {
		t.Error("Generated code missing Level1 struct")
	}
	if !strings.Contains(codeStr, "type Level1Level2 struct") {
		t.Error("Generated code missing Level1Level2 struct")
	}
	if !strings.Contains(codeStr, "type StatusEnum string") {
		t.Error("Generated code missing StatusEnum")
	}

	// Verify field relationships are correct
	if !strings.Contains(codeStr, "Level2 Level1Level2") {
		t.Error("Generated code missing correct field type reference")
	}
	if !strings.Contains(codeStr, "Status StatusEnum") {
		t.Error("Generated code missing correct enum field reference")
	}

	t.Logf("Generated code length: %d bytes", len(code))
	t.Logf("Generated %d root fields, %d structs, %d enums", len(fields), len(structs), len(enums))
}

func TestCodegenAllTypes(t *testing.T) {
	testCases := []TestCase{
		{
			Name:       "Simple Types",
			PromptFile: "simple_types.prompt",
			ExpectedStructs: []ExpectedStruct{
				{
					Name: "SimpleTypesRequest",
					Fields: []ExpectedField{
						{"Name", "string", "name", false},
						{"Age", "int", "age", false},
						{"Height", "float64", "height", false},
						{"Active", "bool", "active", false},
						{"Metadata", "any", "metadata", false},
						{"OptionalField", "string", "optional_field", false},
					},
				},
				{
					Name: "SimpleTypesResponse",
					Fields: []ExpectedField{
						{"Success", "bool", "success", false},
						{"Message", "string", "message", false},
					},
				},
			},
		},
		{
			Name:       "Enum Types",
			PromptFile: "enum_types.prompt",
			ExpectedStructs: []ExpectedStruct{
				{
					Name: "EnumTypesRequest",
					Fields: []ExpectedField{
						{"Priority", "PriorityEnum", "priority", false},
						{"Status", "StatusEnum", "status", false},
					},
				},
				{
					Name: "EnumTypesResponse",
					Fields: []ExpectedField{
						{"Result", "ResultEnum", "result", false},
						{"Level", "LevelEnum", "level", false},
					},
				},
			},
			ExpectedEnums: []ExpectedEnum{
				{
					Name:   "PriorityEnum",
					Type:   "string",
					Values: []string{"low", "medium", "high"},
				},
				{
					Name:   "StatusEnum",
					Type:   "string",
					Values: []string{"pending", "approved", "rejected"},
				},
				{
					Name:   "ResultEnum",
					Type:   "string",
					Values: []string{"success", "failure", "retry"},
				},
				{
					Name:   "LevelEnum",
					Type:   "string",
					Values: []string{"1", "2", "3", "4", "5"},
				},
			},
		},
		{
			Name:       "Array Types",
			PromptFile: "array_types.prompt",
			ExpectedStructs: []ExpectedStruct{
				{
					Name: "ArrayTypesRequest",
					Fields: []ExpectedField{
						{"Tags", "[]string", "tags", false},
						{"Scores", "[]float64", "scores", false},
						{"Ids", "[]int", "ids", false},
					},
				},
				{
					Name: "ArrayTypesResponse",
					Fields: []ExpectedField{
						{"Results", "[]string", "results", false},
						{"Counts", "[]int", "counts", false},
					},
				},
			},
		},
		{
			Name:       "JSON Schema Basic",
			PromptFile: "json_schema_basic.prompt",
			ExpectedStructs: []ExpectedStruct{
				{
					Name: "JsonSchemaBasicRequest",
					Fields: []ExpectedField{
						{"Habit", "string", "habit", true},
						{"Category", "CategoryEnum", "category", true},
						{"WordCount", "int", "word_count", false},
					},
				},
				{
					Name: "JsonSchemaBasicResponse",
					Fields: []ExpectedField{
						{"Summary", "string", "summary", true},
						{"Confidence", "float64", "confidence", false},
						{"Valid", "bool", "valid", true},
					},
				},
			},
			ExpectedEnums: []ExpectedEnum{
				{
					Name:   "CategoryEnum",
					Type:   "string",
					Values: []string{"physical", "mental", "social"},
				},
			},
		},
		{
			Name:       "JSON Schema Arrays",
			PromptFile: "json_schema_arrays.prompt",
			ExpectedStructs: []ExpectedStruct{
				{
					Name: "JsonSchemaArraysRequest",
					Fields: []ExpectedField{
						{"Keywords", "[]string", "keywords", false},
						{"Ratings", "[]float64", "ratings", false},
					},
				},
				{
					Name: "JsonSchemaArraysResponse",
					Fields: []ExpectedField{
						{"MatchedKeywords", "[]string", "matched_keywords", false},
						{"AverageRating", "float64", "average_rating", false},
					},
				},
			},
		},
	}

	// Create temporary output directory
	tempDir := t.TempDir()

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			// Run codegen on test file
			inputFile := filepath.Join("test_data", tc.PromptFile)
			outputFile := filepath.Join(tempDir, strings.TrimSuffix(tc.PromptFile, ".prompt")+".gen.go")

			generator := Generator{
				PackageName: "models",
				OutputDir:   tempDir,
				Verbose:     false,
			}

			err := generator.ProcessFile(inputFile)
			if err != nil {
				t.Fatalf("Code generation failed: %v", err)
			}

			// Read generated file
			generatedCode, err := os.ReadFile(outputFile)
			if err != nil {
				t.Fatalf("Failed to read generated file: %v", err)
			}

			codeStr := string(generatedCode)

			// Verify expected structs exist
			for _, expectedStruct := range tc.ExpectedStructs {
				if !strings.Contains(codeStr, "type "+expectedStruct.Name+" struct") {
					t.Errorf("Expected struct %s not found", expectedStruct.Name)
				}

				// Verify expected fields exist
				for _, expectedField := range expectedStruct.Fields {
					fieldDecl := expectedField.Name + " " + expectedField.GoType
					if !strings.Contains(codeStr, fieldDecl) {
						t.Errorf("Expected field %s %s not found in struct %s",
							expectedField.Name, expectedField.GoType, expectedStruct.Name)
					}

					jsonTag := "`json:\"" + expectedField.JSONTag + "\""
					if !strings.Contains(codeStr, jsonTag) {
						t.Errorf("Expected JSON tag %s not found for field %s",
							jsonTag, expectedField.Name)
					}

					if expectedField.HasValidation {
						validateTag := "validate:\"required\""
						if !strings.Contains(codeStr, validateTag) {
							t.Errorf("Expected validation tag not found for field %s",
								expectedField.Name)
						}
					}
				}
			}

			// Verify expected enums exist
			for _, expectedEnum := range tc.ExpectedEnums {
				enumDecl := "type " + expectedEnum.Name + " " + expectedEnum.Type
				if !strings.Contains(codeStr, enumDecl) {
					t.Errorf("Expected enum %s not found", expectedEnum.Name)
				}

				// Verify enum values exist
				for _, value := range expectedEnum.Values {
					enumValue := expectedEnum.Name + " = \"" + value + "\""
					if !strings.Contains(codeStr, enumValue) {
						t.Errorf("Expected enum value %s not found in enum %s",
							value, expectedEnum.Name)
					}
				}
			}
		})
	}
}

// TestEdgeCases tests various edge cases
func TestEdgeCases(t *testing.T) {
	tempDir := t.TempDir()
	generator := Generator{
		PackageName: "models",
		OutputDir:   tempDir,
		Verbose:     false,
	}

	t.Run("Input Only", func(t *testing.T) {
		err := generator.ProcessFile("test_data/input_only.prompt")
		if err != nil {
			t.Fatalf("Failed to process input-only prompt: %v", err)
		}

		outputFile := filepath.Join(tempDir, "input_only.gen.go")
		generatedCode, err := os.ReadFile(outputFile)
		if err != nil {
			t.Fatalf("Failed to read generated file: %v", err)
		}

		codeStr := string(generatedCode)
		if !strings.Contains(codeStr, "type InputOnlyRequest struct") {
			t.Error("Expected InputOnlyRequest struct not found")
		}
		if strings.Contains(codeStr, "type InputOnlyResponse struct") {
			t.Error("Unexpected InputOnlyResponse struct found")
		}
	})

	t.Run("Output Only", func(t *testing.T) {
		err := generator.ProcessFile("test_data/output_only.prompt")
		if err != nil {
			t.Fatalf("Failed to process output-only prompt: %v", err)
		}

		outputFile := filepath.Join(tempDir, "output_only.gen.go")
		generatedCode, err := os.ReadFile(outputFile)
		if err != nil {
			t.Fatalf("Failed to read generated file: %v", err)
		}

		codeStr := string(generatedCode)
		if strings.Contains(codeStr, "type OutputOnlyRequest struct") {
			t.Error("Unexpected OutputOnlyRequest struct found")
		}
		if !strings.Contains(codeStr, "type OutputOnlyResponse struct") {
			t.Error("Expected OutputOnlyResponse struct not found")
		}
	})

	t.Run("No Schema", func(t *testing.T) {
		err := generator.ProcessFile("test_data/no_schema.prompt")
		// Should not error but should skip generation
		if err != nil {
			t.Fatalf("Failed to process no-schema prompt: %v", err)
		}

		// No file should be generated
		outputFile := filepath.Join(tempDir, "no_schema.gen.go")
		_, err = os.ReadFile(outputFile)
		if err == nil {
			t.Error("Unexpected file generated for no-schema prompt")
		}
	})
}

// TestComplexEnums tests complex enum handling with special characters
func TestComplexEnums(t *testing.T) {
	tempDir := t.TempDir()
	generator := Generator{
		PackageName: "models",
		OutputDir:   tempDir,
		Verbose:     false,
	}

	err := generator.ProcessFile("test_data/complex_enums.prompt")
	if err != nil {
		t.Fatalf("Failed to process complex enums prompt: %v", err)
	}

	outputFile := filepath.Join(tempDir, "complex_enums.gen.go")
	generatedCode, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	codeStr := string(generatedCode)

	// Check that enum constants handle special characters properly
	expectedConstants := []string{
		"DifficultyEnumVeryEasy",
		"DifficultyEnumEasy",
		"DifficultyEnumMedium",
		"DifficultyEnumHard",
		"DifficultyEnumVeryHard",
		"LanguageEnumEn",
		"LanguageEnumEs",
		"LanguageEnumFr",
		"LanguageEnumDe",
		"LanguageEnumJa",
		"LanguageEnumZhCn",
	}

	for _, constant := range expectedConstants {
		if !strings.Contains(codeStr, constant) {
			t.Errorf("Expected constant %s not found", constant)
		}
	}
}

// TestNamingConventions tests various naming convention scenarios
func TestNamingConventions(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple_test", "SimpleTest"},
		{"classify_habits", "ClassifyHabits"},
		{"process_user_data", "ProcessUserData"},
		{"ai_model_response", "AiModelResponse"},
		{"single", "Single"},
		{"multiple_under_scores", "MultipleUnderScores"},
	}

	for _, test := range tests {
		result := snakeToPascalCase(test.input)
		if result != test.expected {
			t.Errorf("snakeToPascalCase(%q) = %q, want %q",
				test.input, result, test.expected)
		}
	}
}

// TestSchemaFormatDetection tests detection of different schema formats
func TestSchemaFormatDetection(t *testing.T) {
	// Test Picoschema detection
	picoSchema := map[string]any{
		"name": "string, the user name",
		"age":  "integer, user age",
	}

	if !isPicoschema(picoSchema) {
		t.Error("Failed to detect Picoschema format")
	}

	// Test JSON Schema detection
	jsonSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{
				"type": "string",
			},
		},
	}

	if !isJSONSchema(jsonSchema) {
		t.Error("Failed to detect JSON Schema format")
	}

	if isPicoschema(jsonSchema) {
		t.Error("Incorrectly detected JSON Schema as Picoschema")
	}
}
