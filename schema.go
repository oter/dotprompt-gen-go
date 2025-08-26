package main

import (
	"fmt"
	"regexp"
	"strings"
)

// Type mapping from schema types to Go types
var picoschemaToGoType = map[string]string{
	"string":  "string",
	"number":  "float64",
	"integer": "int",
	"boolean": "bool",
	"any":     "any",
	"null":    "*any",
}

var jsonSchemaToGoType = map[string]string{
	"string":  "string",
	"number":  "float64",
	"integer": "int",
	"boolean": "bool",
	"array":   "[]", // Will be combined with items type
	// "object" is handled specially in parseJSONSchemaField
}

// ParseSchema parses a schema and returns Go fields and enums (legacy function)
func ParseSchema(schema any, requiredFields []string) ([]GoField, []GoEnum, error) {
	fields, enums, _, err := ParseSchemaWithStructs(schema, requiredFields)
	return fields, enums, err
}

// ParseSchemaWithStructs parses a schema and returns Go fields, enums, and nested structs
func ParseSchemaWithStructs(
	schema any,
	requiredFields []string,
) ([]GoField, []GoEnum, []GoStruct, error) {
	if schema == nil {
		return nil, nil, nil, nil
	}

	// Try to detect schema format and parse accordingly
	if isPicoschema(schema) {
		fields, enums, err := parsePicoschema(schema, requiredFields)
		return fields, enums, nil, err // Picoschema doesn't support nested structs yet
	} else if isJSONSchema(schema) {
		return parseJSONSchemaWithStructs(schema, requiredFields)
	}

	return nil, nil, nil, fmt.Errorf("unsupported schema format")
}

// isPicoschema detects if schema is in Picoschema format
func isPicoschema(schema any) bool {
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

// isJSONSchema detects if schema is in JSON Schema format
func isJSONSchema(schema any) bool {
	schemaMap, ok := schema.(map[string]any)
	if !ok {
		return false
	}

	// JSON Schema has "type" and/or "properties"
	_, hasType := schemaMap["type"]
	_, hasProperties := schemaMap["properties"]

	return hasType || hasProperties
}

// parsePicoschema parses Picoschema format
func parsePicoschema(schema any, requiredFields []string) ([]GoField, []GoEnum, error) {
	schemaMap, ok := schema.(map[string]any)
	if !ok {
		return nil, nil, fmt.Errorf("schema must be an object")
	}

	var fields []GoField
	var enums []GoEnum
	requiredSet := make(map[string]bool)

	for _, field := range requiredFields {
		requiredSet[field] = true
	}

	for fieldName, fieldDef := range schemaMap {
		field, enumDef, err := parsePicoschemaField(fieldName, fieldDef, requiredSet[fieldName])
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

// parseJSONSchemaWithStructs parses JSON Schema format and returns nested structs
func parseJSONSchemaWithStructs(
	schema any,
	requiredFields []string,
) ([]GoField, []GoEnum, []GoStruct, error) {
	schemaMap, ok := schema.(map[string]any)
	if !ok {
		return nil, nil, nil, fmt.Errorf("schema must be an object")
	}

	var fields []GoField
	var enums []GoEnum
	var allStructs []GoStruct
	requiredSet := make(map[string]bool)

	for _, field := range requiredFields {
		requiredSet[field] = true
	}

	// Handle properties
	properties, ok := schemaMap["properties"].(map[string]any)
	if !ok {
		return nil, nil, nil, fmt.Errorf("JSON schema must have properties")
	}

	for fieldName, fieldDef := range properties {
		field, allFieldEnums, directStruct, deeplyNestedStructs, err := parseJSONSchemaFieldWithNestedRecursive(
			fieldName,
			fieldDef,
			requiredSet[fieldName],
			"",
		)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to parse field %s: %w", fieldName, err)
		}

		fields = append(fields, field)
		// Collect all enums from this field (including deeply nested ones)
		if len(allFieldEnums) > 0 {
			enums = append(enums, allFieldEnums...)
		}
		if directStruct != nil {
			allStructs = append(allStructs, *directStruct)
		}
		// Collect all deeply nested structs recursively
		if len(deeplyNestedStructs) > 0 {
			allStructs = append(allStructs, deeplyNestedStructs...)
		}
	}

	return fields, enums, allStructs, nil
}

// getFieldType determines the type of a field string (enum, array, or simple)
func getFieldType(fieldStr string) string {
	if strings.Contains(fieldStr, "(enum") && strings.Contains(fieldStr, "[") &&
		strings.Contains(fieldStr, "]") {
		return "enum"
	}
	if strings.Contains(fieldStr, "(array") && strings.Contains(fieldStr, ":") {
		return "array"
	}
	return "simple"
}

// parsePicoschemaField parses a single field in Picoschema format
func parsePicoschemaField(
	fieldName string,
	fieldDef any,
	isRequired bool,
) (GoField, *GoEnum, error) {
	fieldStr, ok := fieldDef.(string)
	if !ok {
		return GoField{}, nil, fmt.Errorf("picoschema field must be a string")
	}

	// Parse field definition: "type, description" or "type(enum): [values], description"
	field := GoField{
		Name:    schemaFieldToGoField(fieldName),
		JSONTag: fieldName,
	}

	if isRequired {
		field.Validation = "required"
	}

	// Check for optional field marker
	if strings.HasSuffix(fieldName, "?") {
		field.JSONTag = strings.TrimSuffix(fieldName, "?")
		field.Name = schemaFieldToGoField(field.JSONTag)
	}

	// Parse type and description - need to handle enums carefully since they contain commas
	var typeDescPart string
	var description string

	// Check if this is an enum or array definition that contains commas
	fieldType := getFieldType(fieldStr)
	switch fieldType {
	case "enum":
		// Find the end of the enum definition (closing bracket)
		closingBracket := strings.Index(fieldStr, "]")
		if closingBracket != -1 {
			typeDescPart = strings.TrimSpace(fieldStr[:closingBracket+1])
			if closingBracket+1 < len(fieldStr) {
				remaining := strings.TrimSpace(fieldStr[closingBracket+1:])
				if strings.HasPrefix(remaining, ",") {
					description = strings.TrimSpace(remaining[1:])
				}
			}
		} else {
			typeDescPart = fieldStr
		}
	case "array":
		// Handle array definitions
		parts := strings.Split(fieldStr, ",")
		typeDescPart = strings.TrimSpace(parts[0])
		if len(parts) > 1 {
			description = strings.TrimSpace(strings.Join(parts[1:], ","))
		}
	default:
		// Simple type - split on first comma
		parts := strings.SplitN(fieldStr, ",", 2)
		typeDescPart = strings.TrimSpace(parts[0])
		if len(parts) > 1 {
			description = strings.TrimSpace(parts[1])
		}
	}

	field.Comment = description

	// Check for enum definition: "type(enum): [value1, value2]"
	if strings.Contains(typeDescPart, "(enum") {
		return parsePicoschemaEnum(field, typeDescPart)
	}

	// Check for array definition: "fieldname(array): elementType"
	if strings.Contains(fieldName, "(array)") {
		// Clean up field name and JSON tag by removing (array) suffix
		cleanFieldName := strings.Replace(fieldName, "(array)", "", 1)
		field.Name = schemaFieldToGoField(cleanFieldName)
		field.JSONTag = cleanFieldName

		// For array parsing we need the full field string split by comma
		parts := strings.Split(fieldStr, ",")
		return parsePicoschemaArray(field, typeDescPart, parts)
	}

	// Simple type
	field.GoType = mapPicoschemaType(typeDescPart)
	return field, nil, nil
}

// parseJSONSchemaFieldWithNestedRecursive parses a single field and returns all nested structs and
// enums recursively
// parentStructName is used to create unique names for deeply nested structs
func parseJSONSchemaFieldWithNestedRecursive(
	fieldName string,
	fieldDef any,
	isRequired bool,
	parentStructName string,
) (GoField, []GoEnum, *GoStruct, []GoStruct, error) {
	fieldDefMap, ok := fieldDef.(map[string]any)
	if !ok {
		return GoField{}, nil, nil, nil, fmt.Errorf("JSON schema field must be an object")
	}

	field := GoField{
		Name:    schemaFieldToGoField(fieldName),
		JSONTag: fieldName,
	}

	if isRequired {
		field.Validation = "required"
	}

	// Get description
	if desc, ok := fieldDefMap["description"].(string); ok {
		field.Comment = desc
	}

	// Get type
	fieldType, ok := fieldDefMap["type"].(string)
	if !ok {
		field.GoType = "any"
		return field, nil, nil, nil, nil
	}

	// Check for enum
	if enumValues, hasEnum := fieldDefMap["enum"]; hasEnum {
		field, enumDef, err := parseJSONSchemaEnum(field, fieldType, enumValues)
		if err != nil {
			return field, nil, nil, nil, err
		}
		return field, []GoEnum{*enumDef}, nil, nil, nil
	}

	// Check for array
	if fieldType == "array" {
		field = parseJSONSchemaArray(field, fieldDefMap)
		return field, nil, nil, nil, nil
	}

	// Handle object type - create nested struct with unique naming
	if fieldType == "object" {
		// Create unique struct name to avoid conflicts in deeply nested structures
		if parentStructName != "" {
			field.Name = parentStructName + field.Name
		}
		return parseJSONSchemaObjectField(field, fieldDefMap)
	}

	// Simple type
	field.GoType = mapJSONSchemaType(fieldType)
	return field, nil, nil, nil, nil
}

// parseJSONSchemaObjectField handles object type fields by creating nested structs
// Returns the field, any enums (including nested), the main struct, deeply nested structs, and all
// deeply nested enums
func parseJSONSchemaObjectField(
	field GoField,
	fieldDefMap map[string]any,
) (GoField, []GoEnum, *GoStruct, []GoStruct, error) {
	// Create struct name from field name
	structName := field.Name

	// Parse the properties of the nested object
	properties, ok := fieldDefMap["properties"].(map[string]any)
	if !ok {
		// If no properties defined, fall back to map[string]any
		field.GoType = "map[string]any"
		return field, nil, nil, nil, nil
	}

	// Get required fields for this nested object
	var requiredFields []string
	if required, ok := fieldDefMap["required"].([]any); ok {
		for _, req := range required {
			if reqStr, ok := req.(string); ok {
				requiredFields = append(requiredFields, reqStr)
			}
		}
	}

	// Parse nested properties and collect all deeply nested structs and enums
	var nestedFields []GoField
	var allEnums []GoEnum
	var allDeeplyNestedStructs []GoStruct
	requiredSet := make(map[string]bool)
	for _, reqField := range requiredFields {
		requiredSet[reqField] = true
	}

	for propName, propDef := range properties {
		nestedField, allNestedEnums, directNestedStruct, deeplyNestedStructs, err := parseJSONSchemaFieldWithNestedRecursive(
			propName,
			propDef,
			requiredSet[propName],
			structName,
		)
		if err != nil {
			return field, nil, nil, nil, fmt.Errorf(
				"failed to parse nested field %s: %w",
				propName,
				err,
			)
		}

		nestedFields = append(nestedFields, nestedField)

		// Collect all enums (direct and deeply nested)
		if len(allNestedEnums) > 0 {
			allEnums = append(allEnums, allNestedEnums...)
		}

		if directNestedStruct != nil {
			allDeeplyNestedStructs = append(allDeeplyNestedStructs, *directNestedStruct)
		}
		// Collect all deeply nested structs recursively
		if len(deeplyNestedStructs) > 0 {
			allDeeplyNestedStructs = append(allDeeplyNestedStructs, deeplyNestedStructs...)
		}
	}

	// Create the nested struct
	nestedStruct := &GoStruct{
		Name:     structName,
		Fields:   nestedFields,
		Comments: []string{fmt.Sprintf("%s represents %s", structName, field.Comment)},
	}

	// Update field to reference the new struct type
	field.GoType = structName

	return field, allEnums, nestedStruct, allDeeplyNestedStructs, nil
}

// parsePicoschemaEnum parses enum definition in Picoschema
func parsePicoschemaEnum(field GoField, typeDescPart string) (GoField, *GoEnum, error) {
	// Extract enum values from: "string(enum): [value1, value2], description"
	re := regexp.MustCompile(`(\w+)\(enum[^)]*\):\s*\[([^\]]+)\]`)
	matches := re.FindStringSubmatch(typeDescPart)
	if len(matches) != 3 {
		return field, nil, fmt.Errorf("invalid enum format: %s", typeDescPart)
	}

	baseType := matches[1]
	valuesStr := matches[2]

	// Parse enum values
	valueStrs := strings.Split(valuesStr, ",")
	var enumValues []EnumValue

	enumTypeName := field.Name + "Enum"

	for _, valueStr := range valueStrs {
		value := strings.TrimSpace(valueStr)
		constName := enumValueToConstName(enumTypeName, value)
		enumValues = append(enumValues, EnumValue{
			ConstName: constName,
			Value:     value,
		})
	}

	field.GoType = enumTypeName
	field.IsEnum = true

	enum := &GoEnum{
		Name:    enumTypeName,
		Type:    mapPicoschemaType(baseType),
		Values:  enumValues,
		Comment: fmt.Sprintf("valid %s values", field.JSONTag),
	}

	return field, enum, nil
}

// parseJSONSchemaEnum parses enum definition in JSON Schema
func parseJSONSchemaEnum(
	field GoField,
	_ string,
	enumValues any,
) (GoField, *GoEnum, error) {
	enumSlice, ok := enumValues.([]any)
	if !ok {
		return field, nil, fmt.Errorf("enum values must be an array")
	}

	var values []EnumValue
	enumTypeName := field.Name + "Enum"

	for _, val := range enumSlice {
		valueStr := fmt.Sprintf("%v", val)
		constName := enumValueToConstName(enumTypeName, valueStr)
		values = append(values, EnumValue{
			ConstName: constName,
			Value:     valueStr,
		})
	}

	field.GoType = enumTypeName
	field.IsEnum = true

	enum := &GoEnum{
		Name:    enumTypeName,
		Type:    "string", // All enums are string-based in Go for JSON compatibility
		Values:  values,
		Comment: fmt.Sprintf("valid %s values", field.JSONTag),
	}

	return field, enum, nil
}

// parsePicoschemaArray parses array definition in Picoschema
func parsePicoschemaArray(
	field GoField,
	_ string,
	parts []string,
) (GoField, *GoEnum, error) {
	// The typeDescPart is already the element type (e.g., "string")
	// parts[0] contains the element type and parts[1+] contain the description
	if len(parts) < 1 {
		return field, nil, fmt.Errorf("array field missing element type")
	}

	elementType := strings.TrimSpace(parts[0])
	field.GoType = "[]" + mapPicoschemaType(elementType)

	return field, nil, nil
}

// parseJSONSchemaArray parses array definition in JSON Schema
func parseJSONSchemaArray(field GoField, fieldDefMap map[string]any) GoField {
	items, ok := fieldDefMap["items"]
	if !ok {
		field.GoType = "[]any"
		return field
	}

	itemsMap, ok := items.(map[string]any)
	if !ok {
		field.GoType = "[]any"
		return field
	}

	itemType, ok := itemsMap["type"].(string)
	if !ok {
		field.GoType = "[]any"
		return field
	}

	field.GoType = "[]" + mapJSONSchemaType(itemType)
	return field
}

// mapPicoschemaType maps Picoschema types to Go types
func mapPicoschemaType(schemaType string) string {
	if goType, exists := picoschemaToGoType[schemaType]; exists {
		return goType
	}
	return "any" // Fallback
}

// mapJSONSchemaType maps JSON Schema types to Go types
func mapJSONSchemaType(schemaType string) string {
	if goType, exists := jsonSchemaToGoType[schemaType]; exists {
		return goType
	}
	return "any" // Fallback
}
