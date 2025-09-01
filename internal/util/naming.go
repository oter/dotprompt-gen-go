package naming

import (
	"strings"
)

// SnakeToPascalCase converts snake_case to PascalCase
// This is the canonical implementation used throughout the codebase.
func SnakeToPascalCase(s string) string {
	if s == "" {
		return s
	}

	parts := strings.Split(s, "_")

	var result strings.Builder

	for _, part := range parts {
		if len(part) > 0 {
			result.WriteString(strings.ToUpper(part[:1]))

			if len(part) > 1 {
				result.WriteString(part[1:])
			}
		}
	}

	return result.String()
}

// EnumValueToConstName converts an enum value to a Go constant name
// Handles special characters and ensures valid Go identifier.
func EnumValueToConstName(enumTypeName, enumValue string) string {
	// Convert enum value to PascalCase and prefix with type name
	cleanValue := strings.ReplaceAll(enumValue, "-", "_")
	pascalValue := SnakeToPascalCase(cleanValue)

	return enumTypeName + pascalValue
}

// SchemaFieldToGoField converts a schema field name to a Go field name.
func SchemaFieldToGoField(fieldName string) string {
	return SnakeToPascalCase(fieldName)
}
