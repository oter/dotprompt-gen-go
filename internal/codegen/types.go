package codegen

// GoField represents a field in a Go struct.
type GoField struct {
	Name       string
	GoType     string
	JSONTag    string
	Comment    string
	IsEnum     bool
	EnumValues []string
	IsObject   bool // indicates nested struct
	IsPointer  bool // indicates pointer field
}

// NeedsValidation returns true if this field requires validation.
func (f GoField) NeedsValidation() bool {
	return f.IsEnum || f.IsObject
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
func (s GoStruct) NeedsValidation() bool {
	// All structs need validation methods
	return true
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
