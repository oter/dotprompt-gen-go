package prompts

import (
	"testing"
)

func TestImpactLevelEnumValidation(t *testing.T) {
	tests := []struct {
		name    string
		value   ImpactLevelEnum
		wantErr bool
	}{
		{
			name:    "valid growth level",
			value:   ImpactLevelEnumGrowth,
			wantErr: false,
		},
		{
			name:    "valid mastery level",
			value:   ImpactLevelEnumMastery,
			wantErr: false,
		},
		{
			name:    "invalid level",
			value:   "invalid_level",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.value.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("ImpactLevelEnum.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTransformationCategoryEnumValidation(t *testing.T) {
	tests := []struct {
		name    string
		value   TransformationCategoryEnum
		wantErr bool
	}{
		{
			name:    "valid physical vitality",
			value:   TransformationCategoryEnumPhysicalVitality,
			wantErr: false,
		},
		{
			name:    "valid mental mastery",
			value:   TransformationCategoryEnumMentalMastery,
			wantErr: false,
		},
		{
			name:    "invalid category",
			value:   "invalid_category",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.value.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("TransformationCategoryEnum.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUserRoleEnumValidation(t *testing.T) {
	tests := []struct {
		name    string
		value   UserRoleEnum
		wantErr bool
	}{
		{
			name:    "valid admin role",
			value:   UserRoleEnumAdmin,
			wantErr: false,
		},
		{
			name:    "valid user role",
			value:   UserRoleEnumUser,
			wantErr: false,
		},
		{
			name:    "invalid role",
			value:   "invalid_role",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.value.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("UserRoleEnum.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidationTagsExample(t *testing.T) {
	// Test that we can create structs with validation tags
	// The actual validation would be done by external libraries like go-playground/validator

	// This test verifies that the generated structs can be instantiated
	// In real usage, you would use a validation library like:
	// validator := validator.New()
	// err := validator.Struct(input)

	input := ClassifyHabitsOutput{
		Explanation:            "Good habit for growth",
		ImpactLevel:            ImpactLevelEnumGrowth,
		TransformationCategory: TransformationCategoryEnumPhysicalVitality,
	}

	// Verify struct fields are accessible
	if input.Explanation == "" {
		t.Error("Expected explanation to be set")
	}

	// Individual enum validation still works
	if err := input.ImpactLevel.Validate(); err != nil {
		t.Errorf("Expected valid impact level, got error: %v", err)
	}

	if err := input.TransformationCategory.Validate(); err != nil {
		t.Errorf("Expected valid transformation category, got error: %v", err)
	}
}
