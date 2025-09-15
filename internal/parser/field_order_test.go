package parser

import (
	"testing"

	"github.com/oter/dotprompt-gen-go/internal/codegen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFieldOrderPreservation(t *testing.T) {
	tests := []struct {
		name                string
		yamlContent         string
		expectedInputOrder  []string
		expectedOutputOrder []string
	}{
		{
			name: "picoschema field order",
			yamlContent: `model: openai/gpt-4
input:
  schema:
    name: string, the user name
    age: integer, user age
    height: number, user height
    active: boolean, whether user is active
    metadata: any, additional metadata
output:
  schema:
    success: boolean, operation success
    message: string, result message`,
			expectedInputOrder:  []string{"name", "age", "height", "active", "metadata"},
			expectedOutputOrder: []string{"success", "message"},
		},
		{
			name: "json schema field order",
			yamlContent: `model: openai/gpt-4
input:
  schema:
    type: object
    properties:
      first_field:
        type: string
        description: First field
      second_field:
        type: integer
        description: Second field
      third_field:
        type: boolean
        description: Third field
output:
  schema:
    type: object
    properties:
      result:
        type: string
      status:
        type: boolean
      count:
        type: integer`,
			expectedInputOrder:  []string{"first_field", "second_field", "third_field"},
			expectedOutputOrder: []string{"result", "status", "count"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test field order extraction
			inputOrder, err := extractSchemaFieldOrder(tt.yamlContent, "input", "schema")
			require.NoError(t, err)
			assert.Equal(t, tt.expectedInputOrder, inputOrder, "Input field order should match")

			outputOrder, err := extractSchemaFieldOrder(tt.yamlContent, "output", "schema")
			require.NoError(t, err)
			assert.Equal(t, tt.expectedOutputOrder, outputOrder, "Output field order should match")

			// Test that the parsed prompt file includes field order
			promptFile, err := ParsePromptContent("---\n"+tt.yamlContent+"\n---\nTest template", "test.prompt")
			require.NoError(t, err)
			assert.Equal(t, tt.expectedInputOrder, promptFile.InputFieldOrder, "PromptFile input field order should match")
			assert.Equal(t, tt.expectedOutputOrder, promptFile.OutputFieldOrder, "PromptFile output field order should match")
		})
	}
}

func TestSchemaParsingWithFieldOrder(t *testing.T) {
	// Test that schema parsing uses preserved field order
	schema := map[string]any{
		"name":     "string, the user name",
		"age":      "integer, user age",
		"height":   "number, user height",
		"active":   "boolean, whether user is active",
		"metadata": "any, additional metadata",
	}

	fieldOrder := []string{"name", "age", "height", "active", "metadata"}

	fields, enums, err := parsePicoschemaWithFieldOrder(
		schema,
		[]string{},
		SchemaTypeInput,
		fieldOrder,
	)
	require.NoError(t, err)
	require.Len(t, fields, 5)
	require.Len(t, enums, 0) // No enums in this test

	// Verify fields are in the correct order
	expectedFieldNames := []string{"Name", "Age", "Height", "Active", "Metadata"}
	for i, field := range fields {
		assert.Equal(t, expectedFieldNames[i], field.Name, "Field %d should have correct name", i)
	}

	// Test fallback to alphabetical order when no field order provided
	fieldsNoOrder, _, err := parsePicoschemaWithFieldOrder(
		schema,
		[]string{},
		SchemaTypeInput,
		nil, // No field order
	)
	require.NoError(t, err)
	require.Len(t, fieldsNoOrder, 5)

	// Should be in alphabetical order: Active, Age, Height, Metadata, Name
	expectedAlphabeticalNames := []string{"Active", "Age", "Height", "Metadata", "Name"}
	for i, field := range fieldsNoOrder {
		assert.Equal(t, expectedAlphabeticalNames[i], field.Name, "Field %d should be in alphabetical order", i)
	}
}

func TestNestedFieldOrderPreservation(t *testing.T) {
	yamlContent := `model: openai/gpt-4
output:
  schema:
    type: object
    properties:
      user_profile:
        type: object
        properties:
          id:
            type: string
          user_role:
            type: string
            enum: [admin, user, guest]
      recommendations:
        type: array
        items:
          type: string
      success:
        type: boolean`

	// Test nested field order extraction
	nestedFieldOrder, err := extractNestedSchemaFieldOrder(yamlContent, "output", "schema")
	require.NoError(t, err)
	require.NotNil(t, nestedFieldOrder)

	expectedNestedOrder := map[string][]string{
		"user_profile": {"id", "user_role"},
	}
	assert.Equal(t, expectedNestedOrder, nestedFieldOrder, "Nested field order should match schema")

	// Test that the parsed prompt file includes nested field order
	promptFile, err := ParsePromptContent("---\n"+yamlContent+"\n---\nTest template", "test.prompt")
	require.NoError(t, err)
	assert.Equal(t, expectedNestedOrder, promptFile.OutputNestedFieldOrder, "PromptFile nested field order should match")

	// Test that schema parsing uses the nested field order
	_, _, structs, err := ParseJSONSchemaWithNestedFieldOrder(
		promptFile.GetOutputSchema(),
		promptFile.GetRequiredOutputFields(),
		SchemaTypeOutput,
		promptFile.OutputFieldOrder,
		promptFile.OutputNestedFieldOrder,
	)
	require.NoError(t, err)
	require.Len(t, structs, 1) // Should have UserProfile struct

	// Find the UserProfile struct
	var userProfileStruct *codegen.GoStruct
	for i := range structs {
		if structs[i].Name == "UserProfile" {
			userProfileStruct = &structs[i]
			break
		}
	}
	require.NotNil(t, userProfileStruct, "Should have UserProfile struct")

	// Verify field order in the nested struct
	require.Len(t, userProfileStruct.Fields, 2)
	assert.Equal(t, "Id", userProfileStruct.Fields[0].Name, "First field should be Id")
	assert.Equal(t, "UserRole", userProfileStruct.Fields[1].Name, "Second field should be UserRole")
}
