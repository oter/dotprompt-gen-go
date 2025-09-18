package parser

import (
	"errors"
	"sort"
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
	return ParseSchemaWithStructsAndFieldOrder(schema, requiredFields, schemaType, nil)
}

// ParseSchemaWithStructsAndFieldOrder parses a schema with preserved field order.
func ParseSchemaWithStructsAndFieldOrder(
	schema any,
	requiredFields []string,
	schemaType SchemaType,
	fieldOrder []string,
) ([]codegen.GoField, []codegen.GoEnum, []codegen.GoStruct, error) {
	if schema == nil {
		return nil, nil, nil, nil
	}

	// Try to detect schema format and parse accordingly
	if IsPicoschema(schema) {
		fields, enums, err := parsePicoschemaWithFieldOrder(schema, requiredFields, schemaType, fieldOrder)

		return fields, enums, nil, err // Picoschema doesn't support nested structs yet
	} else if IsJSONSchema(schema) {
		return parseJSONSchemaWithStructsAndFieldOrder(schema, requiredFields, schemaType, fieldOrder)
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

// buildOrderedFieldNames creates an ordered list of field names from a schema map.
// Uses preserved field order if available, otherwise falls back to alphabetical sorting.
func buildOrderedFieldNames(schemaFields map[string]any, fieldOrder []string) []string {
	var fieldNames []string

	if len(fieldOrder) > 0 {
		// Use preserved order, but only include fields that exist in schema
		for _, fieldName := range fieldOrder {
			if _, exists := schemaFields[fieldName]; exists {
				fieldNames = append(fieldNames, fieldName)
			}
		}

		// Add any remaining fields not in the preserved order (edge case)
		for fieldName := range schemaFields {
			found := false
			for _, orderedField := range fieldNames {
				if orderedField == fieldName {
					found = true

					break
				}
			}
			if !found {
				fieldNames = append(fieldNames, fieldName)
			}
		}
	} else {
		// Fallback to alphabetical sorting for consistency
		for fieldName := range schemaFields {
			fieldNames = append(fieldNames, fieldName)
		}
		sort.Strings(fieldNames)
	}

	return fieldNames
}

// buildRequiredFieldsSet creates a set of required fields based on schema type.
// For input schemas, all fields are required. For output schemas, use provided required fields.
func buildRequiredFieldsSet(schemaFields map[string]any, requiredFields []string, schemaType SchemaType) map[string]bool {
	requiredSet := make(map[string]bool)

	if schemaType == SchemaTypeInput {
		// All fields are treated as required for input schemas
		for fieldName := range schemaFields {
			requiredSet[fieldName] = true
		}
	} else {
		// For output schemas, use the provided required fields list
		for _, field := range requiredFields {
			requiredSet[field] = true
		}
	}

	return requiredSet
}
