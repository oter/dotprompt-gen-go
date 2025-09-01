package template

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/aymerick/raymond"
)

const (
	// expectedPathRegexGroups is the expected number of regex capture groups for path expressions.
	expectedPathRegexGroups = 3
)

// ValidationResult contains the result of template validation.
type ValidationResult struct {
	Valid        bool
	Errors       []ValidationError
	Warnings     []ValidationError
	Variables    []string
	Helpers      []HelperUsage
	BlockHelpers []BlockHelperUsage
}

// ValidationError represents a template validation error.
type ValidationError struct {
	Message string
	Line    int
	Column  int
	Type    string // "syntax", "variable", "helper"
}

// HelperUsage represents usage of a helper function.
type HelperUsage struct {
	Name       string
	Parameters []string
	Line       int
	Column     int
}

// BlockHelperUsage represents usage of a block helper.
type BlockHelperUsage struct {
	Name       string
	Parameters []string
	Line       int
	Column     int
}

// ValidateHandlebarsTemplate validates a handlebars template string.
func ValidateHandlebarsTemplate(templateContent string) *ValidationResult {
	result := &ValidationResult{
		Valid:        true,
		Variables:    []string{},
		Helpers:      []HelperUsage{},
		BlockHelpers: []BlockHelperUsage{},
	}

	// Parse the template using raymond
	template, err := raymond.Parse(templateContent)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Message: fmt.Sprintf("Template syntax error: %v", err),
			Type:    "syntax",
		})

		return result
	}

	// Get AST string and parse it
	astString := template.PrintAST()
	parseASTString(astString, result)

	return result
}

// parseASTString parses the AST string output from raymond.PrintAST().
func parseASTString(astString string, result *ValidationResult) {
	lines := strings.Split(astString, "\n")

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse different AST node types
		switch {
		case strings.HasPrefix(line, "{{ PATH:"):
			parsePathExpression(line, result)
		case strings.HasPrefix(line, "{{     PATH:"):
			// Handle variables inside block programs
			parsePathExpression(line, result)
		case strings.HasPrefix(line, "BLOCK:"):
			parseBlockExpression(lines[i:], result)
		}
	}
}

// parsePathExpression parses lines like "{{ PATH:name [] }}" or "{{ PATH:role ["system"] }}".
func parsePathExpression(line string, result *ValidationResult) {
	// Regex to match: {{ PATH:name [params] }} or {{     PATH: [] (for {{this}})
	pathRegex := regexp.MustCompile(`\{\{[\s]*PATH:([^\s\[\]]*)\s*\[([^\]]*)\]`)
	matches := pathRegex.FindStringSubmatch(line)

	if len(matches) != expectedPathRegexGroups {
		return
	}

	pathName := matches[1]
	paramsStr := matches[2]

	// Handle special case of {{this}} which shows up as empty path
	if pathName == "" {
		pathName = "this"
	}

	// Convert path separators (user/name -> user.name)
	pathName = strings.ReplaceAll(pathName, "/", ".")

	// Parse parameters
	var params []string

	if strings.TrimSpace(paramsStr) != "" {
		// Handle different parameter formats
		if strings.Contains(paramsStr, "\"") {
			// String parameters like ["system"]
			stringRegex := regexp.MustCompile(`"([^"]*)"`)

			stringMatches := stringRegex.FindAllStringSubmatch(paramsStr, -1)
			for _, match := range stringMatches {
				if len(match) > 1 {
					params = append(params, match[1])
				}
			}
		} else if after, ok := strings.CutPrefix(paramsStr, "PATH:"); ok {
			// Path parameters like [PATH:items]
			pathParam := after
			params = append(params, pathParam)
		}
	}

	// If there are parameters, it's a helper function
	if len(params) > 0 {
		result.Helpers = append(result.Helpers, HelperUsage{
			Name:       pathName,
			Parameters: params,
		})
	} else {
		// No parameters, it's a variable
		result.Variables = append(result.Variables, pathName)
	}
}

// parseBlockExpression parses block helpers from AST string.
func parseBlockExpression(lines []string, result *ValidationResult) {
	for i, line := range lines {
		line = strings.TrimSpace(line)

		// Look for PATH line after BLOCK:
		if strings.HasPrefix(line, "PATH:") && i > 0 &&
			strings.TrimSpace(lines[i-1]) == "BLOCK:" {
			parseBlockPath(line, result)
		}
	}
}

// parseBlockPath parses block helper path like "PATH:each [PATH:items]".
func parseBlockPath(line string, result *ValidationResult) {
	// Regex to match: PATH:helperName [params]
	pathRegex := regexp.MustCompile(`PATH:([^\s\[]+)\s*\[([^\]]*)\]`)
	matches := pathRegex.FindStringSubmatch(line)

	if len(matches) != expectedPathRegexGroups {
		return
	}

	helperName := matches[1]
	paramsStr := matches[2]

	var params []string

	if strings.TrimSpace(paramsStr) != "" {
		if after, ok := strings.CutPrefix(paramsStr, "PATH:"); ok {
			// Path parameter like PATH:items
			pathParam := after
			params = append(params, pathParam)
			// Also add as a variable since it's referenced
			result.Variables = append(result.Variables, pathParam)
		}
	}

	result.BlockHelpers = append(result.BlockHelpers, BlockHelperUsage{
		Name:       helperName,
		Parameters: params,
	})
}

// ValidateVariablesAgainstSchema validates that template variables exist in the schema.
func ValidateVariablesAgainstSchema(variables []string, schema map[string]any) []ValidationError {
	var errors []ValidationError

	// Extract schema properties
	schemaProps := make(map[string]bool)
	if properties, ok := schema["properties"].(map[string]any); ok {
		for prop := range properties {
			schemaProps[prop] = true
		}
	}

	// Check each variable against schema
	for _, variable := range variables {
		// Handle nested properties (user.email -> user)
		rootVar := strings.Split(variable, ".")[0]

		// Skip special handlebars variables
		if isSpecialVariable(rootVar) {
			continue
		}

		if !schemaProps[rootVar] {
			errors = append(errors, ValidationError{
				Message: fmt.Sprintf("Variable '%s' not found in input schema", variable),
				Type:    "variable",
			})
		}
	}

	return errors
}

// isSpecialVariable checks if a variable is a special handlebars variable.
func isSpecialVariable(variable string) bool {
	specialVars := map[string]bool{
		"this":   true,
		"@index": true,
		"@key":   true,
		"@first": true,
		"@last":  true,
	}

	return specialVars[variable]
}

// ValidateHelpers validates helper function usage.
func ValidateHelpers(helpers []HelperUsage, blockHelpers []BlockHelperUsage) []ValidationError {
	var errors []ValidationError

	// Validate regular helpers
	for _, helper := range helpers {
		switch helper.Name {
		case "role":
			errors = append(errors, validateRoleHelper(helper)...)
		default:
			// Unknown helper - could be a warning
			errors = append(errors, ValidationError{
				Message: fmt.Sprintf("Unknown helper function '%s'", helper.Name),
				Type:    "helper",
			})
		}
	}

	// Validate block helpers
	for _, blockHelper := range blockHelpers {
		switch blockHelper.Name {
		case "each":
			errors = append(errors, validateEachHelper(blockHelper)...)
		case "if", "unless":
			errors = append(errors, validateConditionalHelper(blockHelper)...)
		default:
			errors = append(errors, ValidationError{
				Message: fmt.Sprintf("Unknown block helper '%s'", blockHelper.Name),
				Type:    "helper",
			})
		}
	}

	return errors
}

// validateRoleHelper validates the {{role}} helper.
func validateRoleHelper(helper HelperUsage) []ValidationError {
	var errors []ValidationError

	if len(helper.Parameters) != 1 {
		errors = append(errors, ValidationError{
			Message: fmt.Sprintf("role helper expects 1 parameter, got %d", len(helper.Parameters)),
			Type:    "helper",
		})

		return errors
	}

	validRoles := map[string]bool{
		"system":    true,
		"user":      true,
		"assistant": true,
	}

	role := helper.Parameters[0]
	if !validRoles[role] {
		errors = append(errors, ValidationError{
			Message: fmt.Sprintf("Invalid role '%s'. Valid roles: system, user, assistant", role),
			Type:    "helper",
		})
	}

	return errors
}

// validateEachHelper validates the {{#each}} block helper.
func validateEachHelper(blockHelper BlockHelperUsage) []ValidationError {
	var errors []ValidationError

	if len(blockHelper.Parameters) == 0 {
		errors = append(errors, ValidationError{
			Message: "each helper requires a collection parameter",
			Type:    "helper",
		})
	}

	return errors
}

// validateConditionalHelper validates {{#if}} and {{#unless}} helpers.
func validateConditionalHelper(blockHelper BlockHelperUsage) []ValidationError {
	var errors []ValidationError

	if len(blockHelper.Parameters) == 0 {
		errors = append(errors, ValidationError{
			Message: blockHelper.Name + " helper requires a condition parameter",
			Type:    "helper",
		})
	}

	return errors
}
