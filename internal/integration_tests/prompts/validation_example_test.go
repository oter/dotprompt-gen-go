package prompts

import (
	"testing"

	"github.com/oter/dotprompt-gen-go/pkg/validator"
)

func TestClassifyHabitsOutputValidation(t *testing.T) {
	tests := []struct {
		name    string
		output  ClassifyHabitsOutput
		wantErr bool
	}{
		{
			name: "valid output",
			output: ClassifyHabitsOutput{
				Explanation:            "Good habit for growth",
				ImpactLevel:            ImpactLevelEnumGrowth,
				TransformationCategory: TransformationCategoryEnumPhysicalVitality,
			},
			wantErr: false,
		},
		{
			name: "invalid impact level",
			output: ClassifyHabitsOutput{
				Explanation:            "Test explanation",
				ImpactLevel:            "invalid_level",
				TransformationCategory: TransformationCategoryEnumPhysicalVitality,
			},
			wantErr: true,
		},
		{
			name: "invalid transformation category",
			output: ClassifyHabitsOutput{
				Explanation:            "Test explanation",
				ImpactLevel:            ImpactLevelEnumGrowth,
				TransformationCategory: "invalid_category",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.output.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("ClassifyHabitsOutput.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMixedFormatsOutputValidation(t *testing.T) {
	validUserRole := UserRoleEnumAdmin

	tests := []struct {
		name    string
		output  MixedFormatsOutput
		wantErr bool
	}{
		{
			name: "valid output with valid user profile",
			output: MixedFormatsOutput{
				Success: true,
				UserProfile: UserProfile{
					Id:       &[]string{"user123"}[0],
					UserRole: &validUserRole,
				},
			},
			wantErr: false,
		},
		{
			name: "invalid user role in nested struct",
			output: MixedFormatsOutput{
				Success: true,
				UserProfile: UserProfile{
					Id:       &[]string{"user123"}[0],
					UserRole: (*UserRoleEnum)(&[]string{"invalid_role"}[0]),
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.output.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("MixedFormatsOutput.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatorUtilityFunctions(t *testing.T) {
	// Test ValidateAll with multiple outputs
	output1 := ClassifyHabitsOutput{
		Explanation:            "Good habit",
		ImpactLevel:            ImpactLevelEnumGrowth,
		TransformationCategory: TransformationCategoryEnumPhysicalVitality,
	}

	output2 := MixedFormatsOutput{
		Success: true,
		UserProfile: UserProfile{
			UserRole: (*UserRoleEnum)(&[]UserRoleEnum{UserRoleEnumUser}[0]),
		},
	}

	// Test valid outputs
	err := validator.ValidateAll(output1, output2)
	if err != nil {
		t.Errorf("ValidateAll() with valid outputs should not error, got: %v", err)
	}

	// Test with invalid output
	invalidOutput := ClassifyHabitsOutput{
		ImpactLevel: "invalid",
	}

	err = validator.ValidateAll(output1, invalidOutput)
	if err == nil {
		t.Error("ValidateAll() with invalid output should error")
	}
}
