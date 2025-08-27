package internal_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oter/dotprompt-gen-go/internal/generator"
	"github.com/oter/dotprompt-gen-go/internal/model"
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

// TestCodegenAllTypes tests all supported types with end-to-end generation
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

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			// Create separate temporary directory for each test case
			tempDir := t.TempDir()

			// Run codegen on test file
			inputFile := filepath.Join("testdata", tc.PromptFile)
			outputFile := filepath.Join(tempDir, strings.TrimSuffix(tc.PromptFile, ".prompt")+".gen.go")

			gen := model.Generator{
				PackageName: "models",
				OutputDir:   tempDir,
				Verbose:     false,
			}

			err := generator.ProcessFile(gen, inputFile)
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
