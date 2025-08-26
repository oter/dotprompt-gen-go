package main

import (
	"path/filepath"
	"strings"
)

// FilenameToStructNames converts a filename to Go struct names
func FilenameToStructNames(filename string) (request, response string) {
	base := strings.TrimSuffix(filepath.Base(filename), ".prompt")

	// Convert snake_case to PascalCase
	pascal := snakeToPascalCase(base)

	return pascal + "Request", pascal + "Response"
}

// snakeToPascalCase converts snake_case to PascalCase
func snakeToPascalCase(s string) string {
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

// schemaFieldToGoField converts a schema field name to a Go field name
func schemaFieldToGoField(fieldName string) string {
	return snakeToPascalCase(fieldName)
}

// enumValueToConstName converts an enum value to a Go constant name
func enumValueToConstName(enumTypeName, enumValue string) string {
	// Convert enum value to PascalCase and prefix with type name
	cleanValue := strings.ReplaceAll(enumValue, "-", "_")
	pascalValue := snakeToPascalCase(cleanValue)
	return enumTypeName + pascalValue
}
