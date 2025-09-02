package integration_tests

import (
	"testing"

	"github.com/oter/dotprompt-gen-go/internal/integration_tests/prompts"
	"github.com/stretchr/testify/assert"
)

// TestEnumValidationRuntime tests that validation methods work correctly at runtime
func TestEnumValidationRuntime(t *testing.T) {
	// Test PriorityEnum with valid values
	validPriorities := []prompts.PriorityEnum{prompts.PriorityEnumLow, prompts.PriorityEnumMedium, prompts.PriorityEnumHigh}
	for _, p := range validPriorities {
		assert.NoError(t, p.Validate(), "Valid priority %q failed validation", p)
	}

	// Test PriorityEnum with invalid values
	invalidPriorities := []prompts.PriorityEnum{"", "invalid", "LOW", "super-high"}
	for _, p := range invalidPriorities {
		err := p.Validate()
		assert.Error(t, err, "Invalid priority %q passed validation", p)
		if err != nil {
			// Check error message contains the invalid value
			assert.Contains(t, err.Error(), string(p), "Error message doesn't contain invalid value %q", p)
			// Check error message lists valid values
			assert.Contains(t, err.Error(), "low, medium, high", "Error message doesn't list valid values")
		}
	}

	// Test StatusEnum with valid values
	validStatuses := []prompts.StatusEnum{prompts.StatusEnumPending, prompts.StatusEnumApproved, prompts.StatusEnumRejected}
	for _, s := range validStatuses {
		assert.NoError(t, s.Validate(), "Valid status %q failed validation", s)
	}

	// Test StatusEnum with invalid values
	invalidStatuses := []prompts.StatusEnum{"", "unknown", "PENDING", "disabled"}
	for _, s := range invalidStatuses {
		err := s.Validate()
		assert.Error(t, err, "Invalid status %q passed validation", s)
		if err != nil {
			// Check error message format
			assert.Contains(t, err.Error(), string(s), "Error message doesn't contain invalid value %q", s)
			assert.Contains(t, err.Error(), "pending, approved, rejected", "Error message doesn't list valid values")
		}
	}

	// Test ResultEnum with valid values
	validResults := []prompts.ResultEnum{prompts.ResultEnumSuccess, prompts.ResultEnumFailure, prompts.ResultEnumRetry}
	for _, r := range validResults {
		assert.NoError(t, r.Validate(), "Valid result %q failed validation", r)
	}

	// Test ResultEnum with invalid values
	invalidResults := []prompts.ResultEnum{"", "unknown", "SUCCESS", "failed"}
	for _, r := range invalidResults {
		err := r.Validate()
		assert.Error(t, err, "Invalid result %q passed validation", r)
		if err != nil {
			// Check error message format contains both invalid value and valid options
			errStr := err.Error()
			assert.Contains(t, errStr, string(r), "Error message doesn't contain invalid value %q", r)
			assert.Contains(t, errStr, "success, failure, retry", "Error message doesn't list valid values")
		}
	}

	// Test ConfidenceLevelEnum (numeric values)
	validLevels := []prompts.ConfidenceLevelEnum{prompts.ConfidenceLevelEnum1, prompts.ConfidenceLevelEnum2, prompts.ConfidenceLevelEnum3, prompts.ConfidenceLevelEnum4, prompts.ConfidenceLevelEnum5}
	for _, l := range validLevels {
		assert.NoError(t, l.Validate(), "Valid level %q failed validation", l)
	}

	// Test ConfidenceLevelEnum with invalid values
	invalidLevels := []prompts.ConfidenceLevelEnum{"", "0", "6", "ten", "1.0"}
	for _, l := range invalidLevels {
		err := l.Validate()
		assert.Error(t, err, "Invalid level %q passed validation", l)
		if err != nil {
			// Check error message format
			errStr := err.Error()
			assert.Contains(t, errStr, string(l), "Error message doesn't contain invalid value %q", l)
			assert.Contains(t, errStr, "1, 2, 3, 4, 5", "Error message doesn't list valid values")
		}
	}

	// Test UserStatusEnum with valid values
	validUserStatuses := []prompts.UserStatusEnum{prompts.UserStatusEnumActive, prompts.UserStatusEnumInactive, prompts.UserStatusEnumSuspended}
	for _, s := range validUserStatuses {
		assert.NoError(t, s.Validate(), "Valid user status %q failed validation", s)
	}

	// Test UserStatusEnum with invalid values
	invalidUserStatuses := []prompts.UserStatusEnum{"", "unknown", "ACTIVE", "disabled"}
	for _, s := range invalidUserStatuses {
		err := s.Validate()
		assert.Error(t, err, "Invalid user status %q passed validation", s)
		if err != nil {
			// Check error message format
			assert.Contains(t, err.Error(), string(s), "Error message doesn't contain invalid value %q", s)
			assert.Contains(t, err.Error(), "active, inactive, suspended", "Error message doesn't list valid values")
		}
	}
}
