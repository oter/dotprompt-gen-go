package codegen

// GoField represents a field in a Go struct.
type GoField struct {
	Name       string
	GoType     string
	JSONTag    string
	Comment    string
	IsEnum     bool
	EnumValues []string
	IsObject   bool              // indicates nested struct
	IsPointer  bool              // indicates pointer field
	ExtraTags  map[string]string // additional struct tags (e.g., validate:"required")
}

// NeedsValidation returns true if this field requires validation.
func (f GoField) NeedsValidation() bool {
	return f.IsEnum || f.IsObject
}

// StructTags returns the complete struct tag string for this field.
func (f GoField) StructTags() string {
	var tags []string

	// Check if user provided a custom json tag in ExtraTags
	hasCustomJSON := false
	for tagName := range f.ExtraTags {
		if tagName == "json" {
			hasCustomJSON = true

			break
		}
	}

	// Add default JSON tag only if no custom one is provided
	if !hasCustomJSON {
		tags = append(tags, `json:"`+f.JSONTag+`"`)
	}

	// Add all extra tags
	for tagName, tagValue := range f.ExtraTags {
		tags = append(tags, tagName+`:"`+tagValue+`"`)
	}

	if len(tags) == 1 {
		return tags[0]
	}

	// Join multiple tags with spaces
	result := ""
	for i, tag := range tags {
		if i > 0 {
			result += " "
		}
		result += tag
	}

	return result
}

// GoStruct represents a Go struct to be generated.
type GoStruct struct {
	Name     string    // Struct identifier
	Comments []string  // Documentation describing the struct
	Fields   []GoField // Core struct fields
	Enums    []GoEnum  // Related enums
	IsInput  bool      // explicitly mark input structs
	IsOutput bool      // explicitly mark output structs
}

// HasValidationFields returns true if this struct has any fields requiring validation.
func (s GoStruct) HasValidationFields() bool {
	for _, field := range s.Fields {
		if field.NeedsValidation() {
			return true
		}
	}

	return false
}

// NeedsValidation returns true if this struct needs a master Validate() method.
// Only enums need Validate() methods now - structs use validation tags instead.
func (s GoStruct) NeedsValidation() bool {
	return false
}

// GoEnum represents a Go enum/constant type.
type GoEnum struct {
	Name    string      // Enum identifier
	Comment string      // Documentation describing the enum
	Type    string      // Underlying type (string, int, etc.)
	Values  []EnumValue // Enum values
}

// EnumValue represents a single enum value.
type EnumValue struct {
	ConstName string
	Value     string
}

// TemplateData represents data passed to Go code template.
type TemplateData struct {
	Version string     // Used in generated file header
	Package string     // Go file package declaration
	Imports []string   // Go file imports section
	Enums   []GoEnum   // Enum types with receiver functions
	Structs []GoStruct // Struct types with receiver functions
}

// Generator holds configuration for code generation.
type Generator struct {
	PackageName string
	OutputDir   string
	Verbose     bool
}
