package internal_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oter/dotprompt-gen-go/internal/generator"
	"github.com/oter/dotprompt-gen-go/internal/model"
)

// TestEdgeCases tests various edge cases
func TestEdgeCases(t *testing.T) {
	t.Run("Input Only", func(t *testing.T) {
		// Create separate temporary directory for each test case
		tempDir := t.TempDir()
		gen := model.Generator{
			PackageName: "models",
			OutputDir:   tempDir,
			Verbose:     false,
		}
		err := generator.ProcessFile(gen, "testdata/input_only.prompt")
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
		// Create separate temporary directory for each test case
		tempDir := t.TempDir()
		gen := model.Generator{
			PackageName: "models",
			OutputDir:   tempDir,
			Verbose:     false,
		}

		err := generator.ProcessFile(gen, "testdata/output_only.prompt")
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
		// Create separate temporary directory for each test case
		tempDir := t.TempDir()
		gen := model.Generator{
			PackageName: "models",
			OutputDir:   tempDir,
			Verbose:     false,
		}

		err := generator.ProcessFile(gen, "testdata/no_schema.prompt")
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
