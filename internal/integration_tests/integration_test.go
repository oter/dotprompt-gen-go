package integration_tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oter/dotprompt-gen-go/internal/codegen"
	"github.com/oter/dotprompt-gen-go/internal/generator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	Name    string
	GoType  string
	JSONTag string
}

// ExpectedEnum represents expected enum generation
type ExpectedEnum struct {
	Name   string
	Type   string
	Values []string
}

// TestCodegenAllTypes tests all supported types with end-to-end generation
func TestCodegenAllTypes(t *testing.T) {
	testCases := []TestCase{
		{
			Name:       "Simple Types",
			PromptFile: "simple_types.prompt",
			ExpectedStructs: []ExpectedStruct{
				{
					Name: "SimpleTypesInput",
					Fields: []ExpectedField{
						{"Name", "string", "name"},
						{"Age", "int", "age"},
						{"Height", "float64", "height"},
						{"Active", "bool", "active"},
						{"Metadata", "any", "metadata"},
						{"OptionalField", "string", "optional_field"},
					},
				},
				{
					Name: "SimpleTypesOutput",
					Fields: []ExpectedField{
						{"Success", "*bool", "success"},
						{"Message", "*string", "message"},
					},
				},
			},
		},
		{
			Name:       "Comprehensive Enums",
			PromptFile: "comprehensive_enums.prompt",
			ExpectedStructs: []ExpectedStruct{
				{
					Name: "ComprehensiveEnumsInput",
					Fields: []ExpectedField{
						{"Priority", "PriorityEnum", "priority"},
						{"Status", "StatusEnum", "status"},
					},
				},
				{
					Name: "ComprehensiveEnumsOutput",
					Fields: []ExpectedField{
						{"Result", "ResultEnum", "result"},
						{"QualityScore", "QualityScoreEnum", "quality_score"},
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
					Name:   "ConfidenceLevelEnum",
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
					Name: "ArrayTypesInput",
					Fields: []ExpectedField{
						{"Tags", "[]string", "tags"},
						{"Scores", "[]float64", "scores"},
						{"Ids", "[]int", "ids"},
					},
				},
				{
					Name: "ArrayTypesOutput",
					Fields: []ExpectedField{
						{"Results", "[]string", "results"},
						{"Counts", "[]int", "counts"},
					},
				},
			},
		},
		{
			Name:       "JSON Schema Basic",
			PromptFile: "json_schema_basic.prompt",
			ExpectedStructs: []ExpectedStruct{
				{
					Name: "JsonSchemaBasicInput",
					Fields: []ExpectedField{
						{"Habit", "string", "habit"},
						{"HabitCategory", "HabitCategoryEnum", "habit_category"},
						{"WordCount", "int", "word_count"},
					},
				},
				{
					Name: "JsonSchemaBasicOutput",
					Fields: []ExpectedField{
						{"Summary", "string", "summary"},
						{"Confidence", "*float64", "confidence"},
						{"Valid", "bool", "valid"},
					},
				},
			},
			ExpectedEnums: []ExpectedEnum{
				{
					Name:   "HabitCategoryEnum",
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
					Name: "JsonSchemaArraysInput",
					Fields: []ExpectedField{
						{"Keywords", "[]string", "keywords"},
						{"Ratings", "[]float64", "ratings"},
					},
				},
				{
					Name: "JsonSchemaArraysOutput",
					Fields: []ExpectedField{
						{"MatchedKeywords", "[]string", "matched_keywords"},
						{"AverageRating", "*float64", "average_rating"},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			// Create separate temporary directory for each test case
			tempDir := t.TempDir()

			// Run codegen on test file
			inputFile := filepath.Join("prompts", tc.PromptFile)
			outputFile := filepath.Join(tempDir, strings.TrimSuffix(tc.PromptFile, ".prompt")+".gen.go")

			gen := codegen.Generator{
				PackageName: "models",
				OutputDir:   tempDir,
				Verbose:     false,
			}

			err := generator.ProcessFile(gen, inputFile)
			require.NoError(t, err, "Code generation failed")

			// Read generated file
			generatedCode, err := os.ReadFile(outputFile)
			require.NoError(t, err, "Failed to read generated file")

			codeStr := string(generatedCode)

			// Verify expected structs exist
			for _, expectedStruct := range tc.ExpectedStructs {
				assert.Contains(t, codeStr, "type "+expectedStruct.Name+" struct", "Expected struct %s not found", expectedStruct.Name)

				// Verify expected fields exist
				for _, expectedField := range expectedStruct.Fields {
					fieldDecl := expectedField.Name + " " + expectedField.GoType
					assert.Contains(t, codeStr, fieldDecl, "Expected field %s %s not found in struct %s",
						expectedField.Name, expectedField.GoType, expectedStruct.Name)

					jsonTag := "`json:\"" + expectedField.JSONTag + "\""
					assert.Contains(t, codeStr, jsonTag, "Expected JSON tag %s not found for field %s",
						jsonTag, expectedField.Name)
				}
			}

			// Verify expected enums exist
			for _, expectedEnum := range tc.ExpectedEnums {
				enumDecl := "type " + expectedEnum.Name + " " + expectedEnum.Type
				assert.Contains(t, codeStr, enumDecl, "Expected enum %s not found", expectedEnum.Name)

				// Verify enum values exist
				for _, value := range expectedEnum.Values {
					enumValue := expectedEnum.Name + " = \"" + value + "\""
					assert.Contains(t, codeStr, enumValue, "Expected enum value %s not found in enum %s",
						value, expectedEnum.Name)
				}
			}
		})
	}
}
