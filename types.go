package main

// PromptFile represents a parsed dotprompt file
type PromptFile struct {
	Frontmatter FrontmatterData `yaml:",inline"`
	Template    string
	Filename    string
}

// FrontmatterData represents the YAML frontmatter in a dotprompt file
type FrontmatterData struct {
	Model  string     `yaml:"model"`
	Input  SchemaSpec `yaml:"input"`
	Output SchemaSpec `yaml:"output"`
	Config any        `yaml:"config"`
}

// SchemaSpec represents input/output schema specification
type SchemaSpec struct {
	Schema   any      `yaml:"schema"`
	Default  any      `yaml:"default"`
	Required []string `yaml:"required"`
}

// GoField represents a field in a Go struct
type GoField struct {
	Name       string
	GoType     string
	JSONTag    string
	Comment    string
	Validation string
	IsEnum     bool
	EnumValues []string
}

// GoStruct represents a Go struct to be generated
type GoStruct struct {
	Name     string
	Fields   []GoField
	Enums    []GoEnum
	Comments []string
}

// GoEnum represents a Go enum/constant type
type GoEnum struct {
	Name    string
	Type    string
	Values  []EnumValue
	Comment string
}

// EnumValue represents a single enum value
type EnumValue struct {
	ConstName string
	Value     string
}

// TemplateData represents data passed to Go code template
type TemplateData struct {
	Package string
	Structs []GoStruct
	Enums   []GoEnum
	Imports []string
}

// Generator holds configuration for code generation
type Generator struct {
	PackageName string
	OutputDir   string
	Verbose     bool
}
