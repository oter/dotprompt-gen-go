package parser

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/oter/dotprompt-gen-go/internal/codegen"
	"github.com/oter/dotprompt-gen-go/internal/naming"
)

// getJSONSchemaToGoTypeMap returns type mapping from JSON schema types to Go types.
func getJSONSchemaToGoTypeMap() map[string]string {
	return map[string]string{
		"string":  "string",
		"number":  "float64",
		"integer": "int",
		"boolean": "bool",
		"array":   "[]", // Will be combined with items type
		// "object" is handled specially in parseJSONSchemaField
	}
}

// IsJSONSchema detects if schema is in JSON Schema format.
func IsJSONSchema(schema any) bool {
	schemaMap, ok := schema.(map[string]any)
	if !ok {
		return false
	}

	// JSON Schema has "type" and/or "properties"
	_, hasType := schemaMap["type"]
	_, hasProperties := schemaMap["properties"]

	return hasType || hasProperties
}

// ParseJSONSchemaWithNestedFieldOrder parses JSON Schema with nested field order preservation.
func ParseJSONSchemaWithNestedFieldOrder(
	schema any,
	requiredFields []string,
	schemaType SchemaType,
	fieldOrder []string,
	nestedFieldOrder map[string][]string,
) ([]codegen.GoField, []codegen.GoEnum, []codegen.GoStruct, error) {
	return parseJSONSchemaWithStructsAndFieldOrderAndNested(schema, requiredFields, schemaType, fieldOrder, nestedFieldOrder)
}

// parseJSONSchemaWithStructsAndFieldOrder parses JSON Schema format with preserved field order.
func parseJSONSchemaWithStructsAndFieldOrder(
	schema any,
	requiredFields []string,
	schemaType SchemaType,
	fieldOrder []string,
) ([]codegen.GoField, []codegen.GoEnum, []codegen.GoStruct, error) {
	return parseJSONSchemaWithStructsAndFieldOrderAndNested(schema, requiredFields, schemaType, fieldOrder, nil)
}

// parseJSONSchemaWithStructsAndFieldOrderAndNested parses JSON Schema format with preserved field order and nested field order.
func parseJSONSchemaWithStructsAndFieldOrderAndNested(
	schema any,
	requiredFields []string,
	schemaType SchemaType,
	fieldOrder []string,
	nestedFieldOrder map[string][]string,
) ([]codegen.GoField, []codegen.GoEnum, []codegen.GoStruct, error) {
	schemaMap, ok := schema.(map[string]any)
	if !ok {
		return nil, nil, nil, errors.New("schema must be an object")
	}

	var (
		fields     []codegen.GoField
		enums      []codegen.GoEnum
		allStructs []codegen.GoStruct
	)

	// Handle properties first to get field names for input schema logic
	properties, ok := schemaMap["properties"].(map[string]any)
	if !ok {
		return nil, nil, nil, errors.New("JSON schema must have properties")
	}

	// Build required fields set and ordered field names using shared functions
	requiredSet := buildRequiredFieldsSet(properties, requiredFields, schemaType)
	fieldNames := buildOrderedFieldNames(properties, fieldOrder)

	// Process fields in sorted order
	for _, fieldName := range fieldNames {
		fieldDef := properties[fieldName]

		field, allFieldEnums, directStruct, deeplyNestedStructs, err := parseJSONSchemaFieldWithNestedRecursive(
			fieldName,
			fieldDef,
			requiredSet[fieldName],
			"",
			schemaType,
			nestedFieldOrder,
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

// parseJSONSchemaFieldWithNestedRecursive parses a single field and returns all nested structs and
// enums recursively
// parentStructName is used to create unique names for deeply nested structs.
func parseJSONSchemaFieldWithNestedRecursive(
	fieldName string,
	fieldDef any,
	isRequired bool,
	parentStructName string,
	schemaType SchemaType,
	nestedFieldOrder map[string][]string,
) (codegen.GoField, []codegen.GoEnum, *codegen.GoStruct, []codegen.GoStruct, error) {
	fieldDefMap, ok := fieldDef.(map[string]any)
	if !ok {
		return codegen.GoField{}, nil, nil, nil, errors.New("JSON schema field must be an object")
	}

	field := createBaseField(fieldName, isRequired, fieldDefMap)
	fieldType := getFieldTypeFromSchema(fieldDefMap)

	// Handle different field types
	switch {
	case hasEnum(fieldDefMap):
		return handleEnumField(field, fieldType, fieldDefMap, isRequired, schemaType)
	case fieldType == "array":
		return handleArrayField(field, fieldDefMap, isRequired, schemaType)
	case fieldType == "object":
		return handleObjectField(field, fieldDefMap, parentStructName, schemaType, nestedFieldOrder)
	default:
		return handleSimpleField(field, fieldType, isRequired, schemaType)
	}
}

// createBaseField creates a base GoField with common properties.
func createBaseField(fieldName string, _ bool, fieldDefMap map[string]any) codegen.GoField {
	field := codegen.GoField{
		Name:      naming.SchemaFieldToGoField(fieldName),
		JSONTag:   fieldName,
		ExtraTags: make(map[string]string),
	}

	// Get description
	if desc, ok := fieldDefMap["description"].(string); ok {
		field.Comment = desc
	}

	// Parse x-codegen-extra-tags extension
	if extraTags, ok := fieldDefMap["x-codegen-extra-tags"].(map[string]any); ok {
		for tagName, tagValue := range extraTags {
			if tagValueStr, ok := tagValue.(string); ok {
				field.ExtraTags[tagName] = tagValueStr
			}
		}
	}

	return field
}

// getFieldTypeFromSchema extracts the type from schema definition.
func getFieldTypeFromSchema(fieldDefMap map[string]any) string {
	fieldType, ok := fieldDefMap["type"].(string)
	if !ok {
		return "any"
	}

	return fieldType
}

// hasEnum checks if the field definition contains an enum.
func hasEnum(fieldDefMap map[string]any) bool {
	_, hasEnum := fieldDefMap["enum"]

	return hasEnum
}

// handleEnumField processes enum field types.
func handleEnumField(
	field codegen.GoField,
	fieldType string,
	fieldDefMap map[string]any,
	isRequired bool,
	schemaType SchemaType,
) (codegen.GoField, []codegen.GoEnum, *codegen.GoStruct, []codegen.GoStruct, error) {
	enumValues := fieldDefMap["enum"]

	field, enumDef, err := parseJSONSchemaEnum(field, fieldType, enumValues)
	if err != nil {
		return field, nil, nil, nil, err
	}

	// For output schemas, make non-required enum fields pointers
	if schemaType == SchemaTypeOutput && !isRequired {
		field.GoType = "*" + field.GoType
	}

	return field, []codegen.GoEnum{*enumDef}, nil, nil, nil
}

// handleArrayField processes array field types.
func handleArrayField(
	field codegen.GoField,
	fieldDefMap map[string]any,
	_ bool,
	schemaType SchemaType,
) (codegen.GoField, []codegen.GoEnum, *codegen.GoStruct, []codegen.GoStruct, error) {
	// Check if array items are objects with properties
	items, hasItems := fieldDefMap["items"]
	if !hasItems {
		// Handle simple arrays (primitives, etc.)
		field = parseJSONSchemaArray(field, fieldDefMap)

		return field, nil, nil, nil, nil
	}

	itemsMap, isMap := items.(map[string]any)
	if !isMap {
		// Handle simple arrays (primitives, etc.)
		field = parseJSONSchemaArray(field, fieldDefMap)

		return field, nil, nil, nil, nil
	}

	itemType, hasType := itemsMap["type"].(string)
	_, hasProperties := itemsMap["properties"].(map[string]any)
	_, hasEnum := itemsMap["enum"]

	// If items are objects with properties, create a nested struct
	if hasType && itemType == "object" && hasProperties {
		return handleObjectArrayField(field, itemsMap, schemaType)
	}

	// If items have enum values, create an enum type for the array items
	if hasEnum {
		updatedField, enumDef, err := parseJSONSchemaArrayEnum(field, itemsMap)
		if err != nil {
			return field, nil, nil, nil, err
		}

		return updatedField, []codegen.GoEnum{*enumDef}, nil, nil, nil
	}

	// Handle simple arrays (primitives, etc.)
	field = parseJSONSchemaArray(field, fieldDefMap)

	// Arrays are already nillable in Go, no need for pointer types

	return field, nil, nil, nil, nil
}

// handleObjectArrayField processes array field types with object items.
func handleObjectArrayField(
	field codegen.GoField,
	itemsMap map[string]any,
	schemaType SchemaType,
) (codegen.GoField, []codegen.GoEnum, *codegen.GoStruct, []codegen.GoStruct, error) {
	// Create struct name for the array item type
	itemStructName := field.Name + "Item"

	// Create a field representing the object item type
	itemField := codegen.GoField{
		Name:    itemStructName,
		JSONTag: field.JSONTag + "_item", // This won't be used but needed for the struct creation
	}

	// Get description for the item struct
	if desc, ok := itemsMap["description"].(string); ok {
		itemField.Comment = desc
	} else {
		itemField.Comment = fmt.Sprintf("item in %s array", field.JSONTag)
	}

	// Parse the object as if it were a direct object field
	_, allEnums, directStruct, nestedStructs, err := parseJSONSchemaObjectField(
		itemField,
		itemsMap,
		schemaType,
		nil, // Array items don't have nested field order preservation yet
	)
	if err != nil {
		return field, nil, nil, nil, fmt.Errorf("failed to parse array item object: %w", err)
	}

	// Set the array field type to use the item struct
	field.GoType = "[]" + itemStructName

	return field, allEnums, directStruct, nestedStructs, nil
}

// handleObjectField processes object field types with nested structs.
func handleObjectField(
	field codegen.GoField,
	fieldDefMap map[string]any,
	parentStructName string,
	schemaType SchemaType,
	nestedFieldOrder map[string][]string,
) (codegen.GoField, []codegen.GoEnum, *codegen.GoStruct, []codegen.GoStruct, error) {
	// Create unique struct name to avoid conflicts in deeply nested structures
	if parentStructName != "" {
		field.Name = parentStructName + field.Name
	}

	return parseJSONSchemaObjectField(field, fieldDefMap, schemaType, nestedFieldOrder)
}

// handleSimpleField processes simple field types.
func handleSimpleField(
	field codegen.GoField,
	fieldType string,
	isRequired bool,
	schemaType SchemaType,
) (codegen.GoField, []codegen.GoEnum, *codegen.GoStruct, []codegen.GoStruct, error) {
	field.GoType = convertJSONSchemaTypeToGo(fieldType)

	// For output schemas, make non-required fields pointers
	// But skip arrays since they're already nillable
	if schemaType == SchemaTypeOutput && !isRequired && !strings.HasPrefix(field.GoType, "[]") {
		field.GoType = "*" + field.GoType
	}

	field.IsPointer = strings.HasPrefix(field.GoType, "*")

	return field, nil, nil, nil, nil
}

// parseJSONSchemaObjectField handles object type fields by creating nested structs
// Returns the field, any enums (including nested), the main struct, deeply nested structs, and all
// deeply nested enums.
func parseJSONSchemaObjectField(
	field codegen.GoField,
	fieldDefMap map[string]any,
	schemaType SchemaType,
	nestedFieldOrder map[string][]string,
) (codegen.GoField, []codegen.GoEnum, *codegen.GoStruct, []codegen.GoStruct, error) {
	structName := field.Name

	properties, ok := fieldDefMap["properties"].(map[string]any)
	if !ok {
		field.GoType = "map[string]any"

		return field, nil, nil, nil, nil
	}

	requiredFields := extractRequiredFields(fieldDefMap)
	propNames := getOrderedPropertyNames(properties, field.JSONTag, nestedFieldOrder)

	nestedFields, allEnums, allDeeplyNestedStructs, err := processNestedProperties(
		properties, propNames, requiredFields, structName, schemaType, nestedFieldOrder,
	)
	if err != nil {
		return field, nil, nil, nil, err
	}

	nestedStruct := createNestedStruct(structName, field.Comment, nestedFields)
	field = updateFieldForStruct(field, structName)

	return field, allEnums, nestedStruct, allDeeplyNestedStructs, nil
}

// extractRequiredFields extracts required field names from field definition map.
func extractRequiredFields(fieldDefMap map[string]any) []string {
	var requiredFields []string

	if required, ok := fieldDefMap["required"].([]any); ok {
		for _, req := range required {
			if reqStr, ok := req.(string); ok {
				requiredFields = append(requiredFields, reqStr)
			}
		}
	}

	return requiredFields
}

// getOrderedPropertyNames returns property names in the correct order.
func getOrderedPropertyNames(
	properties map[string]any,
	fieldJSONTag string,
	nestedFieldOrder map[string][]string,
) []string {
	fieldOrderForThisStruct := nestedFieldOrder[fieldJSONTag]
	if len(fieldOrderForThisStruct) > 0 {
		return getPreservedOrderPropertyNames(properties, fieldOrderForThisStruct)
	}

	return getAlphabeticalPropertyNames(properties)
}

// getPreservedOrderPropertyNames returns property names using preserved field order.
func getPreservedOrderPropertyNames(properties map[string]any, fieldOrder []string) []string {
	var propNames []string

	// Use preserved order, but only include fields that exist in properties
	for _, fieldName := range fieldOrder {
		if _, exists := properties[fieldName]; exists {
			propNames = append(propNames, fieldName)
		}
	}

	// Add any remaining fields not in the preserved order (edge case)
	for propName := range properties {
		found := false
		for _, orderedField := range propNames {
			if orderedField == propName {
				found = true

				break
			}
		}
		if !found {
			propNames = append(propNames, propName)
		}
	}

	return propNames
}

// getAlphabeticalPropertyNames returns property names in alphabetical order.
func getAlphabeticalPropertyNames(properties map[string]any) []string {
	var propNames []string
	for propName := range properties {
		propNames = append(propNames, propName)
	}
	sort.Strings(propNames)

	return propNames
}

// processNestedProperties processes all nested properties and returns collected data.
func processNestedProperties(
	properties map[string]any,
	propNames []string,
	requiredFields []string,
	structName string,
	schemaType SchemaType,
	nestedFieldOrder map[string][]string,
) ([]codegen.GoField, []codegen.GoEnum, []codegen.GoStruct, error) {
	var (
		nestedFields           []codegen.GoField
		allEnums               []codegen.GoEnum
		allDeeplyNestedStructs []codegen.GoStruct
	)

	requiredSet := make(map[string]bool)
	for _, reqField := range requiredFields {
		requiredSet[reqField] = true
	}

	for _, propName := range propNames {
		propDef := properties[propName]

		nestedField, allNestedEnums, directNestedStruct, deeplyNestedStructs, err := parseJSONSchemaFieldWithNestedRecursive(
			propName,
			propDef,
			requiredSet[propName],
			structName,
			schemaType,
			nestedFieldOrder,
		)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to parse nested field %s: %w", propName, err)
		}

		nestedFields = append(nestedFields, nestedField)
		allEnums = append(allEnums, allNestedEnums...)

		if directNestedStruct != nil {
			allDeeplyNestedStructs = append(allDeeplyNestedStructs, *directNestedStruct)
		}
		allDeeplyNestedStructs = append(allDeeplyNestedStructs, deeplyNestedStructs...)
	}

	return nestedFields, allEnums, allDeeplyNestedStructs, nil
}

// createNestedStruct creates a new nested struct with the given parameters.
func createNestedStruct(structName, comment string, fields []codegen.GoField) *codegen.GoStruct {
	return &codegen.GoStruct{
		Name:     structName,
		Comments: []string{fmt.Sprintf("%s represents %s", structName, comment)},
		Fields:   fields,
	}
}

// updateFieldForStruct updates the field to reference the struct type.
func updateFieldForStruct(field codegen.GoField, structName string) codegen.GoField {
	field.GoType = structName
	field.IsObject = true
	field.IsPointer = strings.HasPrefix(field.GoType, "*")

	return field
}

// parseJSONSchemaEnum parses enum definition in JSON Schema.
func parseJSONSchemaEnum(
	field codegen.GoField,
	_ string,
	enumValues any,
) (codegen.GoField, *codegen.GoEnum, error) {
	enumSlice, ok := enumValues.([]any)
	if !ok {
		return field, nil, errors.New("enum values must be an array")
	}

	var values []codegen.EnumValue

	enumTypeName := field.Name + "Enum"

	for _, val := range enumSlice {
		valueStr := fmt.Sprintf("%v", val)
		constName := naming.EnumValueToConstName(enumTypeName, valueStr)
		values = append(values, codegen.EnumValue{
			ConstName: constName,
			Value:     valueStr,
		})
	}

	field.GoType = enumTypeName
	field.IsEnum = true
	field.IsPointer = strings.HasPrefix(field.GoType, "*")

	enum := &codegen.GoEnum{
		Name:    enumTypeName,
		Comment: fmt.Sprintf("valid %s values", field.JSONTag),
		Type:    "string", // All enums are string-based in Go for JSON compatibility
		Values:  values,
	}

	return field, enum, nil
}

// parseJSONSchemaArrayEnum parses array items with enum values and generates enum type for array.
func parseJSONSchemaArrayEnum(
	field codegen.GoField,
	itemsMap map[string]any,
) (codegen.GoField, *codegen.GoEnum, error) {
	enumValues := itemsMap["enum"]

	enumSlice, ok := enumValues.([]any)
	if !ok {
		return field, nil, errors.New("enum values must be an array")
	}

	var values []codegen.EnumValue

	// Create enum type name for array items
	enumTypeName := field.Name + "ItemEnum"

	for _, val := range enumSlice {
		valueStr := fmt.Sprintf("%v", val)
		constName := naming.EnumValueToConstName(enumTypeName, valueStr)
		values = append(values, codegen.EnumValue{
			ConstName: constName,
			Value:     valueStr,
		})
	}

	// Set array field to use enum type
	field.GoType = "[]" + enumTypeName
	field.IsEnum = false // The field itself is not enum, but array of enums
	field.IsPointer = false

	enum := &codegen.GoEnum{
		Name:    enumTypeName,
		Comment: fmt.Sprintf("valid %s item values", field.JSONTag),
		Type:    "string", // All enums are string-based in Go for JSON compatibility
		Values:  values,
	}

	return field, enum, nil
}

// parseJSONSchemaArray parses array definition in JSON Schema.
func parseJSONSchemaArray(field codegen.GoField, fieldDefMap map[string]any) codegen.GoField {
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

	field.GoType = "[]" + convertJSONSchemaTypeToGo(itemType)

	return field
}

// convertJSONSchemaTypeToGo maps JSON Schema types to Go types.
func convertJSONSchemaTypeToGo(schemaType string) string {
	if goType, exists := getJSONSchemaToGoTypeMap()[schemaType]; exists {
		return goType
	}

	return "any" // Fallback
}
