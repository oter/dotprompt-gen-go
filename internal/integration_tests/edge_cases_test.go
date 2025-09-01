package integration_tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/oter/dotprompt-gen-go/internal/codegen"
	"github.com/oter/dotprompt-gen-go/internal/generator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEdgeCases tests various edge cases
func TestEdgeCases(t *testing.T) {
	t.Run("Input Only", func(t *testing.T) {
		// Create separate temporary directory for each test case
		tempDir := t.TempDir()
		gen := codegen.Generator{
			PackageName: "models",
			OutputDir:   tempDir,
			Verbose:     false,
		}
		err := generator.ProcessFile(gen, filepath.Join("prompts", "input_only.prompt"))
		require.NoError(t, err, "Failed to process input-only prompt")

		outputFile := filepath.Join(tempDir, "input_only.gen.go")
		generatedCode, err := os.ReadFile(outputFile)
		require.NoError(t, err, "Failed to read generated file")

		codeStr := string(generatedCode)
		assert.Contains(t, codeStr, "type InputOnlyInput struct", "Expected InputOnlyInput struct not found")
		assert.NotContains(t, codeStr, "type InputOnlyOutput struct", "Unexpected InputOnlyOutput struct found")
	})

	t.Run("Output Only", func(t *testing.T) {
		// Create separate temporary directory for each test case
		tempDir := t.TempDir()
		gen := codegen.Generator{
			PackageName: "models",
			OutputDir:   tempDir,
			Verbose:     false,
		}

		err := generator.ProcessFile(gen, filepath.Join("prompts", "output_only.prompt"))
		require.NoError(t, err, "Failed to process output-only prompt")

		outputFile := filepath.Join(tempDir, "output_only.gen.go")
		generatedCode, err := os.ReadFile(outputFile)
		require.NoError(t, err, "Failed to read generated file")

		codeStr := string(generatedCode)
		assert.NotContains(t, codeStr, "type OutputOnlyInput struct", "Unexpected OutputOnlyInput struct found")
		assert.Contains(t, codeStr, "type OutputOnlyOutput struct", "Expected OutputOnlyOutput struct not found")
	})

	t.Run("No Schema", func(t *testing.T) {
		// Create separate temporary directory for each test case
		tempDir := t.TempDir()
		gen := codegen.Generator{
			PackageName: "models",
			OutputDir:   tempDir,
			Verbose:     false,
		}

		err := generator.ProcessFile(gen, filepath.Join("prompts", "no_schema.prompt"))
		// Should not error but should skip generation
		require.NoError(t, err, "Failed to process no-schema prompt")

		// No file should be generated
		outputFile := filepath.Join(tempDir, "no_schema.gen.go")
		_, err = os.ReadFile(outputFile)
		assert.Error(t, err, "Unexpected file generated for no-schema prompt")
	})
}
