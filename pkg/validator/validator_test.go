package validator

import (
	"errors"
	"testing"
)

// Mock validator for testing
type mockValidator struct {
	err error
}

func (m mockValidator) Validate() error {
	return m.err
}

func TestValidateAll(t *testing.T) {
	tests := []struct {
		name        string
		validators  []Validator
		wantErr     bool
		errContains string
	}{
		{
			name:       "all valid",
			validators: []Validator{mockValidator{nil}, mockValidator{nil}},
			wantErr:    false,
		},
		{
			name:        "one invalid",
			validators:  []Validator{mockValidator{nil}, mockValidator{errors.New("invalid value")}},
			wantErr:     true,
			errContains: "invalid value",
		},
		{
			name:        "multiple invalid",
			validators:  []Validator{mockValidator{errors.New("error1")}, mockValidator{errors.New("error2")}},
			wantErr:     true,
			errContains: "error1; error2",
		},
		{
			name:        "with nil validator",
			validators:  []Validator{mockValidator{nil}, nil, mockValidator{errors.New("error")}},
			wantErr:     true,
			errContains: "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAll(tt.validators...)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.errContains != "" && err != nil && err.Error() != "validation failed: "+tt.errContains {
				t.Errorf("ValidateAll() error = %v, want to contain %v", err.Error(), tt.errContains)
			}
		})
	}
}

func TestValidateFields(t *testing.T) {
	tests := []struct {
		name        string
		fields      map[string]Validator
		wantErr     bool
		errContains string
	}{
		{
			name:    "all valid fields",
			fields:  map[string]Validator{"field1": mockValidator{nil}, "field2": mockValidator{nil}},
			wantErr: false,
		},
		{
			name:        "one invalid field",
			fields:      map[string]Validator{"field1": mockValidator{nil}, "field2": mockValidator{errors.New("bad value")}},
			wantErr:     true,
			errContains: "field2: bad value",
		},
		{
			name:    "with nil field validator",
			fields:  map[string]Validator{"field1": mockValidator{nil}, "field2": nil},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFields(tt.fields)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFields() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.errContains != "" && err != nil {
				if err.Error() != "validation failed: "+tt.errContains {
					t.Errorf("ValidateFields() error = %q, want to contain %q", err.Error(), "validation failed: "+tt.errContains)
				}
			}
		})
	}
}
