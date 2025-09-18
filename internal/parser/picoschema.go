package parser

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/oter/dotprompt-gen-go/internal/codegen"
	"github.com/oter/dotprompt-gen-go/internal/naming"
)

const (
	// expectedRegexMatchGroups is the expected number of regex capture groups including the full match.
	expectedRegexMatchGroups = 3
)

// getPicoschemaToGoTypeMap returns type mapping from picoschema types to Go types.
func getPicoschemaToGoTypeMap() map[string]string {
	return map[string]string{
		"string":  "string",
		"number":  "float64",
		"integer": "int",
		"boolean": "bool",
		"any":     "any",
		"null":    "*any",
	}
}

// IsPicoschema detects if schema is in Picoschema format.
func IsPicoschema(schema any) bool {
	schemaMap, ok := schema.(map[string]any)
	if !ok {
		return false
	}

	// Picoschema doesn't have "type", "properties" at root level
	// Instead it has field definitions directly
	_, hasType := schemaMap["type"]
	_, hasProperties := schemaMap["properties"]

	return !hasType && !hasProperties
}

// parsePicoschemaWithFieldOrder parses Picoschema format with preserved field order.
func parsePicoschemaWithFieldOrder(
	schema any,
	requiredFields []string,
	schemaType SchemaType,
	fieldOrder []string,
) ([]codegen.GoField, []codegen.GoEnum, error) {
	schemaMap, ok := schema.(map[string]any)
	if !ok {
		return nil, nil, errors.New("schema must be an object")
	}

	var (
		fields []codegen.GoField
		enums  []codegen.GoEnum
	)

	// Build required fields set and ordered field names using shared functions
	requiredSet := buildRequiredFieldsSet(schemaMap, requiredFields, schemaType)
	fieldNames := buildOrderedFieldNames(schemaMap, fieldOrder)

	// Process fields in sorted order
	for _, fieldName := range fieldNames {
		fieldDef := schemaMap[fieldName]

		field, enumDef, err := parsePicoschemaField(
			fieldName,
			fieldDef,
			requiredSet[fieldName],
			schemaType,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse field %s: %w", fieldName, err)
		}

		fields = append(fields, field)
		if enumDef != nil {
			enums = append(enums, *enumDef)
		}
	}

	return fields, enums, nil
}

// parsePicoschemaField parses a single field in Picoschema format.
func parsePicoschemaField(
	fieldName string,
	fieldDef any,
	isRequired bool,
	schemaType SchemaType,
) (codegen.GoField, *codegen.GoEnum, error) {
	fieldStr, ok := fieldDef.(string)
	if !ok {
		return codegen.GoField{}, nil, errors.New("picoschema field must be a string")
	}

	field := createBasePicoschemaField(fieldName)
	typeDescPart, description := parseFieldDefinition(fieldStr)
	field.Comment = description

	// Handle enum definitions
	if strings.Contains(typeDescPart, "(enum") {
		return handlePicoschemaEnum(field, typeDescPart, isRequired, schemaType)
	}

	// Handle array definitions
	if strings.Contains(fieldName, "(array)") {
		return handlePicoschemaArray(field, fieldName, fieldStr, typeDescPart)
	}

	// Handle simple types
	return handleSimplePicoschemaType(field, typeDescPart, isRequired, schemaType), nil, nil
}

// createBasePicoschemaField creates the base field structure.
func createBasePicoschemaField(fieldName string) codegen.GoField {
	field := codegen.GoField{
		Name:      naming.SchemaFieldToGoField(fieldName),
		JSONTag:   fieldName,
		ExtraTags: make(map[string]string),
	}

	// Check for optional field marker
	if strings.HasSuffix(fieldName, "?") {
		field.JSONTag = strings.TrimSuffix(fieldName, "?")
		field.Name = naming.SchemaFieldToGoField(field.JSONTag)
	}

	return field
}

// parseFieldDefinition extracts type and description from field definition.
func parseFieldDefinition(fieldStr string) (string, string) {
	fieldType := detectPicoschemaFieldType(fieldStr)

	switch fieldType {
	case "enum":
		return parseEnumFieldDefinition(fieldStr)
	case "array":
		return parseArrayFieldDefinition(fieldStr)
	default:
		return parseSimpleFieldDefinition(fieldStr)
	}
}

// parseEnumFieldDefinition parses enum field definitions.
func parseEnumFieldDefinition(fieldStr string) (string, string) {
	// Find the end of the enum definition (closing bracket)
	closingBracket := strings.Index(fieldStr, "]")
	if closingBracket == -1 {
		return fieldStr, ""
	}

	typeDescPart := strings.TrimSpace(fieldStr[:closingBracket+1])

	if closingBracket+1 >= len(fieldStr) {
		return typeDescPart, ""
	}

	remaining := strings.TrimSpace(fieldStr[closingBracket+1:])
	if !strings.HasPrefix(remaining, ",") {
		return typeDescPart, ""
	}

	description := strings.TrimSpace(remaining[1:])

	return typeDescPart, description
}

// parseArrayFieldDefinition parses array field definitions.
func parseArrayFieldDefinition(fieldStr string) (string, string) {
	parts := strings.Split(fieldStr, ",")
	typeDescPart := strings.TrimSpace(parts[0])

	if len(parts) <= 1 {
		return typeDescPart, ""
	}

	description := strings.TrimSpace(strings.Join(parts[1:], ","))

	return typeDescPart, description
}

// parseSimpleFieldDefinition parses simple field definitions.
func parseSimpleFieldDefinition(fieldStr string) (string, string) {
	parts := strings.SplitN(fieldStr, ",", 2)
	typeDescPart := strings.TrimSpace(parts[0])

	if len(parts) <= 1 {
		return typeDescPart, ""
	}

	description := strings.TrimSpace(parts[1])

	return typeDescPart, description
}

// handlePicoschemaEnum handles enum field processing.
func handlePicoschemaEnum(
	field codegen.GoField,
	typeDescPart string,
	isRequired bool,
	schemaType SchemaType,
) (codegen.GoField, *codegen.GoEnum, error) {
	field, enumDef, err := parsePicoschemaEnum(field, typeDescPart)
	if err != nil {
		return field, enumDef, err
	}

	// For output schemas, make non-required enum fields pointers
	if schemaType == SchemaTypeOutput && !isRequired {
		field.GoType = "*" + field.GoType
	}

	return field, enumDef, err
}

// handlePicoschemaArray handles array field processing.
func handlePicoschemaArray(
	field codegen.GoField,
	fieldName string,
	fieldStr string,
	typeDescPart string,
) (codegen.GoField, *codegen.GoEnum, error) {
	// Clean up field name and JSON tag by removing (array) suffix
	cleanFieldName := strings.Replace(fieldName, "(array)", "", 1)
	field.Name = naming.SchemaFieldToGoField(cleanFieldName)
	field.JSONTag = cleanFieldName

	// For array parsing we need the full field string split by comma
	parts := strings.Split(fieldStr, ",")

	field, err := parsePicoschemaArray(field, typeDescPart, parts)
	if err != nil {
		return field, nil, err
	}

	// Arrays are already nillable in Go, no need for pointer types
	return field, nil, err
}

// handleSimplePicoschemaType handles simple type processing.
func handleSimplePicoschemaType(
	field codegen.GoField,
	typeDescPart string,
	isRequired bool,
	schemaType SchemaType,
) codegen.GoField {
	field.GoType = convertPicoschemaTypeToGo(typeDescPart)

	// For output schemas, make non-required fields pointers
	// But skip arrays since they're already nillable
	if schemaType == SchemaTypeOutput && !isRequired && !strings.HasPrefix(field.GoType, "[]") {
		field.GoType = "*" + field.GoType
	}

	field.IsPointer = strings.HasPrefix(field.GoType, "*")

	return field
}

// parsePicoschemaEnum parses enum definition in Picoschema.
func parsePicoschemaEnum(
	field codegen.GoField,
	typeDescPart string,
) (codegen.GoField, *codegen.GoEnum, error) {
	// Extract enum values from: "string(enum): [value1, value2], description"
	re := regexp.MustCompile(`(\w+)\(enum[^)]*\):\s*\[([^\]]+)\]`)

	matches := re.FindStringSubmatch(typeDescPart)
	if len(matches) != expectedRegexMatchGroups {
		return field, nil, fmt.Errorf("invalid enum format: %s", typeDescPart)
	}

	baseType := matches[1]
	valuesStr := matches[2]

	// Parse enum values
	valueStrs := strings.Split(valuesStr, ",")

	var enumValues []codegen.EnumValue

	enumTypeName := field.Name + "Enum"

	for _, valueStr := range valueStrs {
		value := strings.TrimSpace(valueStr)
		constName := naming.EnumValueToConstName(enumTypeName, value)
		enumValues = append(enumValues, codegen.EnumValue{
			ConstName: constName,
			Value:     value,
		})
	}

	field.GoType = enumTypeName
	field.IsEnum = true
	field.IsPointer = strings.HasPrefix(field.GoType, "*")

	enum := &codegen.GoEnum{
		Name:    enumTypeName,
		Comment: fmt.Sprintf("valid %s values", field.JSONTag),
		Type:    convertPicoschemaTypeToGo(baseType),
		Values:  enumValues,
	}

	return field, enum, nil
}

// parsePicoschemaArray parses array definition in Picoschema.
func parsePicoschemaArray(
	field codegen.GoField,
	_ string,
	parts []string,
) (codegen.GoField, error) {
	// The typeDescPart is already the element type (e.g., "string")
	// parts[0] contains the element type and parts[1+] contain the description
	if len(parts) < 1 {
		return field, errors.New("array field missing element type")
	}

	elementType := strings.TrimSpace(parts[0])
	field.GoType = "[]" + convertPicoschemaTypeToGo(elementType)
	field.IsPointer = false // Arrays are not pointer types

	return field, nil
}

// convertPicoschemaTypeToGo maps Picoschema types to Go types.
func convertPicoschemaTypeToGo(schemaType string) string {
	if goType, exists := getPicoschemaToGoTypeMap()[schemaType]; exists {
		return goType
	}

	return "any" // Fallback
}
