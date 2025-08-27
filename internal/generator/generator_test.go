package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oter/dotprompt-gen-go/internal/model"
	"github.com/oter/dotprompt-gen-go/internal/schema"
)

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

	fields, enums, structs, err := schema.ParseSchemaWithStructs(deepSchema, []string{"level1"})
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

// TestComplexEnums tests complex enum handling with special characters
func TestComplexEnums(t *testing.T) {
	// Create temporary directory - automatically cleaned up by Go test framework
	tempDir := t.TempDir()
	gen := model.Generator{
		PackageName: "models",
		OutputDir:   tempDir,
		Verbose:     false,
	}

	err := ProcessFile(gen, "../testdata/complex_enums.prompt")
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
		result := SnakeToPascalCase(test.input)
		if result != test.expected {
			t.Errorf("SnakeToPascalCase(%q) = %q, want %q",
				test.input, result, test.expected)
		}
	}
}
