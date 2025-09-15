package parser

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/oter/dotprompt-gen-go/internal/ast"
)

const (
	// minimumFrontmatterParts is the minimum number of parts when splitting by --- delimiters.
	minimumFrontmatterParts = 3
)

// ParsePromptFile parses a dotprompt file and returns a PromptFile.
func ParsePromptFile(filePath string) (*ast.PromptFile, error) {
	// Validate and clean the file path to prevent path traversal attacks
	cleanPath := filepath.Clean(filePath)

	// Ensure the file has the .prompt extension
	if !strings.HasSuffix(cleanPath, ".prompt") {
		return nil, fmt.Errorf("invalid file extension: expected .prompt file, got %s", cleanPath)
	}

	// Convert to absolute path to further validate
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute path for %s: %w", cleanPath, err)
	}

	// Check if file exists and is a regular file (not a directory or special file)
	info, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to access file %s: %w", absPath, err)
	}

	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("path is not a regular file: %s", absPath)
	}

	// #nosec G304 - Path has been validated above
	content, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", absPath, err)
	}

	return ParsePromptContent(string(content), absPath)
}

// ParsePromptContent parses dotprompt content and returns a PromptFile.
func ParsePromptContent(content, filename string) (*ast.PromptFile, error) {
	// Split by frontmatter delimiters
	parts := strings.Split(content, "---")
	if len(parts) < minimumFrontmatterParts {
		return nil, errors.New("invalid dotprompt format: missing frontmatter delimiters")
	}

	// Parse YAML frontmatter
	frontmatterContent := strings.TrimSpace(parts[1])

	var frontmatter ast.FrontmatterData

	err := yaml.Unmarshal([]byte(frontmatterContent), &frontmatter)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML frontmatter: %w", err)
	}

	// Extract field order from YAML for schema fields
	inputFieldOrder, err := extractSchemaFieldOrder(frontmatterContent, "input", "schema")
	if err != nil {
		return nil, fmt.Errorf("failed to extract input field order: %w", err)
	}

	outputFieldOrder, err := extractSchemaFieldOrder(frontmatterContent, "output", "schema")
	if err != nil {
		return nil, fmt.Errorf("failed to extract output field order: %w", err)
	}

	// Extract nested field orders recursively
	inputNestedFieldOrder, err := extractNestedSchemaFieldOrder(frontmatterContent, "input", "schema")
	if err != nil {
		return nil, fmt.Errorf("failed to extract input nested field order: %w", err)
	}

	outputNestedFieldOrder, err := extractNestedSchemaFieldOrder(frontmatterContent, "output", "schema")
	if err != nil {
		return nil, fmt.Errorf("failed to extract output nested field order: %w", err)
	}

	// Extract template content (everything after the second ---)
	templateParts := parts[2:]
	template := strings.TrimSpace(strings.Join(templateParts, "---"))

	promptFile := &ast.PromptFile{
		Filename:    filename,
		Frontmatter: frontmatter,
		Template:    template,
	}

	// Store field order information for later use
	if inputFieldOrder != nil {
		promptFile.InputFieldOrder = inputFieldOrder
	}
	if outputFieldOrder != nil {
		promptFile.OutputFieldOrder = outputFieldOrder
	}
	if inputNestedFieldOrder != nil {
		promptFile.InputNestedFieldOrder = inputNestedFieldOrder
	}
	if outputNestedFieldOrder != nil {
		promptFile.OutputNestedFieldOrder = outputNestedFieldOrder
	}

	return promptFile, nil
}

// extractSchemaFieldOrder extracts the field order from YAML schema using yaml.Node.
func extractSchemaFieldOrder(yamlContent, schemaType, schemaKey string) ([]string, error) {
	var node yaml.Node
	if err := yaml.Unmarshal([]byte(yamlContent), &node); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Find the schema node (input.schema or output.schema)
	schemaNode := findSchemaNode(&node, schemaType, schemaKey)
	if schemaNode == nil {
		return nil, nil // No schema found, not an error
	}

	// Extract field names in order from the schema node
	return extractFieldNamesFromNode(schemaNode), nil
}

// findSchemaNode finds the schema node in the YAML tree.
func findSchemaNode(node *yaml.Node, schemaType, schemaKey string) *yaml.Node {
	if node.Kind != yaml.DocumentNode && node.Kind != yaml.MappingNode {
		return nil
	}

	// For document node, check the first child
	if node.Kind == yaml.DocumentNode {
		if len(node.Content) == 0 {
			return nil
		}
		return findSchemaNode(node.Content[0], schemaType, schemaKey)
	}

	// For mapping node, find the schemaType key (e.g., "input" or "output")
	for i := 0; i < len(node.Content); i += 2 {
		if i+1 >= len(node.Content) {
			break
		}

		keyNode := node.Content[i]
		valueNode := node.Content[i+1]

		if keyNode.Value == schemaType && valueNode.Kind == yaml.MappingNode {
			// Found input/output node, now look for schema key
			for j := 0; j < len(valueNode.Content); j += 2 {
				if j+1 >= len(valueNode.Content) {
					break
				}

				schemaKeyNode := valueNode.Content[j]
				schemaValueNode := valueNode.Content[j+1]

				if schemaKeyNode.Value == schemaKey {
					return schemaValueNode
				}
			}
		}
	}

	return nil
}

// extractFieldNamesFromNode extracts field names in order from a YAML mapping node.
func extractFieldNamesFromNode(node *yaml.Node) []string {
	if node.Kind != yaml.MappingNode {
		return nil
	}

	// For JSON Schema, we need to look inside the "properties" field
	propertiesNode := findPropertiesNode(node)
	if propertiesNode != nil {
		return extractFieldNamesFromPropertiesNode(propertiesNode)
	}

	// For Picoschema, extract field names directly
	var fieldNames []string
	for i := 0; i < len(node.Content); i += 2 {
		if i >= len(node.Content) {
			break
		}

		keyNode := node.Content[i]
		if keyNode.Kind == yaml.ScalarNode {
			fieldNames = append(fieldNames, keyNode.Value)
		}
	}

	return fieldNames
}

// findPropertiesNode finds the "properties" node in a JSON schema.
func findPropertiesNode(node *yaml.Node) *yaml.Node {
	if node.Kind != yaml.MappingNode {
		return nil
	}

	for i := 0; i < len(node.Content); i += 2 {
		if i+1 >= len(node.Content) {
			break
		}

		keyNode := node.Content[i]
		valueNode := node.Content[i+1]

		if keyNode.Value == "properties" && valueNode.Kind == yaml.MappingNode {
			return valueNode
		}
	}

	return nil
}

// extractFieldNamesFromPropertiesNode extracts field names from a JSON schema properties node.
func extractFieldNamesFromPropertiesNode(node *yaml.Node) []string {
	if node.Kind != yaml.MappingNode {
		return nil
	}

	var fieldNames []string
	for i := 0; i < len(node.Content); i += 2 {
		if i >= len(node.Content) {
			break
		}

		keyNode := node.Content[i]
		if keyNode.Kind == yaml.ScalarNode {
			fieldNames = append(fieldNames, keyNode.Value)
		}
	}

	return fieldNames
}

// extractNestedSchemaFieldOrder extracts field orders for all nested objects in a schema.
func extractNestedSchemaFieldOrder(yamlContent, schemaType, schemaKey string) (map[string][]string, error) {
	var node yaml.Node
	if err := yaml.Unmarshal([]byte(yamlContent), &node); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Find the schema node (input.schema or output.schema)
	schemaNode := findSchemaNode(&node, schemaType, schemaKey)
	if schemaNode == nil {
		return nil, nil // No schema found, not an error
	}

	// Extract nested field orders recursively
	nestedOrders := make(map[string][]string)
	extractNestedFieldOrdersRecursive(schemaNode, "", nestedOrders)

	if len(nestedOrders) == 0 {
		return nil, nil
	}

	return nestedOrders, nil
}

// extractNestedFieldOrdersRecursive recursively extracts field orders from nested objects.
func extractNestedFieldOrdersRecursive(node *yaml.Node, currentPath string, nestedOrders map[string][]string) {
	if node.Kind != yaml.MappingNode {
		return
	}

	// Look for "properties" node (JSON Schema) or direct field definitions (Picoschema)
	propertiesNode := findPropertiesNode(node)
	if propertiesNode != nil {
		// JSON Schema format - extract field order from properties
		fieldNames := extractFieldNamesFromPropertiesNode(propertiesNode)
		if len(fieldNames) > 0 && currentPath != "" {
			nestedOrders[currentPath] = fieldNames
		}

		// Recursively process nested objects in properties
		for i := 0; i < len(propertiesNode.Content); i += 2 {
			if i+1 >= len(propertiesNode.Content) {
				break
			}

			keyNode := propertiesNode.Content[i]
			valueNode := propertiesNode.Content[i+1]

			if keyNode.Kind == yaml.ScalarNode && valueNode.Kind == yaml.MappingNode {
				fieldName := keyNode.Value
				var nestedPath string
				if currentPath == "" {
					nestedPath = fieldName
				} else {
					nestedPath = currentPath + "." + fieldName
				}

				// Check if this is an object type
				if isObjectTypeNode(valueNode) {
					extractNestedFieldOrdersRecursive(valueNode, nestedPath, nestedOrders)
				}
			}
		}
	} else {
		// Picoschema format - check for nested objects directly
		// For now, Picoschema doesn't support nested objects, so this is a placeholder
		// for future extension
	}
}

// isObjectTypeNode checks if a YAML node represents an object type.
func isObjectTypeNode(node *yaml.Node) bool {
	if node.Kind != yaml.MappingNode {
		return false
	}

	// Check for type: object or presence of properties
	hasObjectType := false
	hasProperties := false

	for i := 0; i < len(node.Content); i += 2 {
		if i+1 >= len(node.Content) {
			break
		}

		keyNode := node.Content[i]
		valueNode := node.Content[i+1]

		if keyNode.Kind == yaml.ScalarNode {
			switch keyNode.Value {
			case "type":
				if valueNode.Kind == yaml.ScalarNode && valueNode.Value == "object" {
					hasObjectType = true
				}
			case "properties":
				if valueNode.Kind == yaml.MappingNode {
					hasProperties = true
				}
			}
		}
	}

	return hasObjectType && hasProperties
}
