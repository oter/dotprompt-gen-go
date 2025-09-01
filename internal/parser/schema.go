package parser

import (
	"errors"
	"strings"

	"github.com/oter/dotprompt-gen-go/internal/codegen"
)

// SchemaType represents whether the schema is for input or output.
type SchemaType string

const (
	SchemaTypeInput  SchemaType = "input"
	SchemaTypeOutput SchemaType = "output"
)

// ParseSchemaWithStructs parses a schema and returns Go fields, enums, and nested structs.
func ParseSchemaWithStructs(
	schema any,
	requiredFields []string,
	schemaType SchemaType,
) ([]codegen.GoField, []codegen.GoEnum, []codegen.GoStruct, error) {
	if schema == nil {
		return nil, nil, nil, nil
	}

	// Try to detect schema format and parse accordingly
	if IsPicoschema(schema) {
		fields, enums, err := parsePicoschema(schema, requiredFields, schemaType)

		return fields, enums, nil, err // Picoschema doesn't support nested structs yet
	} else if IsJSONSchema(schema) {
		return parseJSONSchemaWithStructs(schema, requiredFields, schemaType)
	}

	return nil, nil, nil, errors.New("unsupported schema format")
}

// detectPicoschemaFieldType determines the type of a Picoschema field string (enum, array, or
// simple).
func detectPicoschemaFieldType(fieldStr string) string {
	if strings.Contains(fieldStr, "(enum") && strings.Contains(fieldStr, "[") &&
		strings.Contains(fieldStr, "]") {
		return "enum"
	}

	if strings.Contains(fieldStr, "(array") && strings.Contains(fieldStr, ":") {
		return "array"
	}

	return "simple"
}
