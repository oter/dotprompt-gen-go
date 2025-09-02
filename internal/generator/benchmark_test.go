package generator

import (
	"fmt"
	"testing"

	"github.com/oter/dotprompt-gen-go/internal/codegen"
)

// BenchmarkGenerateGoCode benchmarks Go code generation performance
func BenchmarkGenerateGoCode(b *testing.B) {
	// Create test structs and enums
	structs := []codegen.GoStruct{
		{
			Name: "TestInput",
			Fields: []codegen.GoField{
				{Name: "Name", GoType: "string", JSONTag: "name", Comment: "User name"},
				{Name: "Age", GoType: "int", JSONTag: "age", Comment: "User age"},
				{Name: "Email", GoType: "string", JSONTag: "email", Comment: "User email"},
				{Name: "Active", GoType: "bool", JSONTag: "active", Comment: "User active status"},
				{Name: "Priority", GoType: "PriorityEnum", JSONTag: "priority", Comment: "Task priority", IsEnum: true},
			},
		},
		{
			Name: "TestOutput",
			Fields: []codegen.GoField{
				{Name: "Success", GoType: "bool", JSONTag: "success", Comment: "Operation success"},
				{Name: "Message", GoType: "string", JSONTag: "message", Comment: "Response message"},
				{Name: "Data", GoType: "map[string]any", JSONTag: "data", Comment: "Response data"},
			},
		},
	}

	enums := []codegen.GoEnum{
		{
			Name:    "PriorityEnum",
			Comment: "Task priority levels",
			Type:    "string",
			Values: []codegen.EnumValue{
				{ConstName: "PriorityEnumLow", Value: "low"},
				{ConstName: "PriorityEnumMedium", Value: "medium"},
				{ConstName: "PriorityEnumHigh", Value: "high"},
			},
		},
		{
			Name:    "StatusEnum",
			Comment: "Status values",
			Type:    "string",
			Values: []codegen.EnumValue{
				{ConstName: "StatusEnumPending", Value: "pending"},
				{ConstName: "StatusEnumApproved", Value: "approved"},
				{ConstName: "StatusEnumRejected", Value: "rejected"},
			},
		},
	}

	for b.Loop() {
		_, err := GenerateGoCode(structs, enums, "testpkg")
		if err != nil {
			b.Fatalf("Failed to generate Go code: %v", err)
		}
	}
}

// BenchmarkGenerateGoCodeLarge benchmarks code generation with many structs and enums
func BenchmarkGenerateGoCodeLarge(b *testing.B) {
	// Create a large number of structs and enums
	var structs []codegen.GoStruct
	var enums []codegen.GoEnum

	// Generate 50 structs with 10 fields each
	for i := range 50 {
		var fields []codegen.GoField
		for j := range 10 {
			fields = append(fields, codegen.GoField{
				Name:    fmt.Sprintf("Field%d", j),
				GoType:  "string",
				JSONTag: fmt.Sprintf("field_%d", j),
				Comment: fmt.Sprintf("Field %d description", j),
			})
		}
		structs = append(structs, codegen.GoStruct{
			Name:   fmt.Sprintf("Struct%d", i),
			Fields: fields,
		})
	}

	// Generate 20 enums with 5 values each
	for i := range 20 {
		var values []codegen.EnumValue
		for j := range 5 {
			values = append(values, codegen.EnumValue{
				ConstName: fmt.Sprintf("Enum%dValue%d", i, j),
				Value:     fmt.Sprintf("value_%d", j),
			})
		}
		enums = append(enums, codegen.GoEnum{
			Name:    fmt.Sprintf("Enum%d", i),
			Comment: fmt.Sprintf("Enum %d description", i),
			Type:    "string",
			Values:  values,
		})
	}

	for b.Loop() {
		_, err := GenerateGoCode(structs, enums, "testpkg")
		if err != nil {
			b.Fatalf("Failed to generate large Go code: %v", err)
		}
	}
}

// BenchmarkFilenameToStructNames benchmarks struct name generation
func BenchmarkFilenameToStructNames(b *testing.B) {
	filenames := []string{
		"simple_test.prompt",
		"classify_habits.prompt",
		"process_user_data.prompt",
		"ai_model_response.prompt",
		"very_long_filename_with_many_underscores.prompt",
	}

	for b.Loop() {
		for _, filename := range filenames {
			FilenameToStructNames(filename)
		}
	}
}
