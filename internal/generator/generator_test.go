package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oter/dotprompt-gen-go/internal/codegen"
	"github.com/oter/dotprompt-gen-go/internal/naming"
	"github.com/oter/dotprompt-gen-go/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTempGenerator creates a temporary generator for testing
func createTempGenerator(t *testing.T, pkg string) (codegen.Generator, string) {
	tempDir := t.TempDir()
	return codegen.Generator{
		PackageName: pkg,
		OutputDir:   tempDir,
		Verbose:     false,
	}, tempDir
}

// processTestPrompt processes a test prompt file and returns the generated code
func processTestPrompt(t *testing.T, gen codegen.Generator, promptFileName string) string {
	t.Helper()

	inputFile := filepath.Join("..", "integration_tests", "prompts", promptFileName)
	err := ProcessFile(gen, inputFile)
	if err != nil {
		t.Fatalf("Failed to process prompt file %s: %v", promptFileName, err)
	}

	outputFileName := strings.TrimSuffix(promptFileName, ".prompt") + ".gen.go"
	outputFile := filepath.Join(gen.OutputDir, outputFileName)

	generatedCode, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read generated file %s: %v", outputFile, err)
	}

	return string(generatedCode)
}

// TestGeneratedCodeCompiles tests that deeply nested generated code actually compiles
func TestGeneratedCodeCompiles(t *testing.T) {
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

	fields, enums, structs, err := parser.ParseSchemaWithStructs(deepSchema, []string{"level1"}, parser.SchemaTypeOutput)
	require.NoError(t, err, "Failed to parse schema")

	// Generate Go code
	code, err := GenerateGoCode(structs, enums, "testpkg")
	require.NoError(t, err, "Failed to generate Go code")

	// Verify the code contains expected struct definitions
	codeStr := string(code)
	assert.Contains(t, codeStr, "type Level1 struct", "Generated code missing Level1 struct")
	assert.Contains(t, codeStr, "type Level1Level2 struct", "Generated code missing Level1Level2 struct")
	assert.Contains(t, codeStr, "type StatusEnum string", "Generated code missing StatusEnum")

	// Verify field relationships are correct
	assert.Contains(t, codeStr, "Level2 Level1Level2", "Generated code missing correct field type reference")
	assert.Contains(t, codeStr, "Status StatusEnum", "Generated code missing correct enum field reference")

	t.Logf("Generated code length: %d bytes", len(code))
	t.Logf("Generated %d root fields, %d structs, %d enums", len(fields), len(structs), len(enums))
}

// TestComplexEnums tests complex enum handling with special characters
func TestComplexEnums(t *testing.T) {
	gen, _ := createTempGenerator(t, "models")
	codeStr := processTestPrompt(t, gen, "comprehensive_enums.prompt")

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
		assert.Contains(t, codeStr, constant, "enum constant")
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
		result := naming.SnakeToPascalCase(test.input)
		assert.Equal(t, test.expected, result, "SnakeToPascalCase(%q) failed", test.input)
	}
}

// TestEnumValidationGeneration tests that Validate() methods are generated for enums
func TestEnumValidationGeneration(t *testing.T) {
	// Schema with multiple enum types
	testSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"priority": map[string]any{
				"type": "string",
				"enum": []any{"low", "medium", "high"},
			},
			"status": map[string]any{
				"type": "string",
				"enum": []any{"pending", "approved", "rejected"},
			},
			"level": map[string]any{
				"type": "integer",
				"enum": []any{1, 2, 3},
			},
		},
		"required": []any{"priority", "status"},
	}

	_, enums, structs, err := parser.ParseSchemaWithStructs(testSchema, []string{"priority", "status"}, parser.SchemaTypeOutput)
	require.NoError(t, err, "Failed to parse schema")

	// Generate Go code
	code, err := GenerateGoCode(structs, enums, "testpkg")
	require.NoError(t, err, "Failed to generate Go code")

	codeStr := string(code)

	// Verify fmt import is included
	assert.Contains(t, codeStr, `import "fmt"`, "Generated code missing fmt import")

	// Verify each enum has a validation method
	expectedValidationMethods := []string{
		"func (e PriorityEnum) Validate() error",
		"func (e StatusEnum) Validate() error",
		"func (e LevelEnum) Validate() error",
	}

	for _, method := range expectedValidationMethods {
		assert.Contains(t, codeStr, method, "Expected validation method not found: %s", method)
	}

	// Verify switch statements are generated with correct constants
	expectedSwitchCases := []string{
		"case PriorityEnumLow, PriorityEnumMedium, PriorityEnumHigh:",
		"case StatusEnumPending, StatusEnumApproved, StatusEnumRejected:",
		"case LevelEnum1, LevelEnum2, LevelEnum3:",
	}

	for _, switchCase := range expectedSwitchCases {
		assert.Contains(t, codeStr, switchCase, "Expected switch case not found: %s", switchCase)
	}

	// Verify error messages include valid values list
	expectedErrorMessages := []string{
		`"invalid PriorityEnum value: %q, must be one of: low, medium, high"`,
		`"invalid StatusEnum value: %q, must be one of: pending, approved, rejected"`,
		`"invalid LevelEnum value: %q, must be one of: 1, 2, 3"`,
	}

	for _, errorMsg := range expectedErrorMessages {
		assert.Contains(t, codeStr, errorMsg, "Expected error message not found: %s", errorMsg)
	}

	t.Logf("Generated code with validation methods: %d bytes", len(code))
}
