package validator

import (
	"fmt"
	"strings"
)

// Validator represents any type that can validate itself.
type Validator interface {
	Validate() error
}

// ValidateAll validates multiple objects implementing Validator interface.
func ValidateAll(validators ...Validator) error {
	var errors []string

	for _, v := range validators {
		if v == nil {
			continue // skip nil validators
		}
		if err := v.Validate(); err != nil {
			errors = append(errors, err.Error())
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}

// ValidateFields validates struct fields by name for better error messages.
func ValidateFields(fieldValidations map[string]Validator) error {
	var errors []string

	for fieldName, validator := range fieldValidations {
		if validator == nil {
			continue
		}
		if err := validator.Validate(); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %s", fieldName, err.Error()))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}
