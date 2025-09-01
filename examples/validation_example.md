# Validation System Example

The `dotprompt-gen-go` tool automatically generates comprehensive validation methods for all generated Output structs.

## Features

- **Master Validate() methods** for all Output structs
- **Individual validation** for enum and nested object fields
- **Nested struct validation** - nested objects get their own Validate() methods
- **Utility functions** for validating multiple objects at once

## Example Generated Code

From this prompt file:

```yaml
---
output:
  schema:
    type: object
    properties:
      user_profile:
        type: object
        properties:
          user_role:
            type: string
            enum: [admin, user, guest]
      status:
        type: string
        enum: [active, inactive]
    required:
      - user_profile
      - status
---
```

Generates:

```go
// Output struct with master validation
type ExampleOutput struct {
    UserProfile UserProfile `json:"user_profile" validate:"required"`
    Status      StatusEnum  `json:"status" validate:"required"`
}

// Master validation method
func (s ExampleOutput) Validate() error {
    fieldValidations := map[string]validator.Validator{
        "userprofile": s.UserProfile,
        "status":      s.Status,
    }
    return validator.ValidateFields(fieldValidations)
}

// Nested struct also gets validation
func (s UserProfile) Validate() error {
    fieldValidations := map[string]validator.Validator{
        "userrole": s.UserRole,
    }
    return validator.ValidateFields(fieldValidations)
}

// Individual enum validation
func (e StatusEnum) Validate() error {
    switch e {
    case StatusEnumActive, StatusEnumInactive:
        return nil
    default:
        return fmt.Errorf("invalid StatusEnum value: %q", string(e))
    }
}
```

## Usage Examples

### Individual Validation

```go
output := ExampleOutput{
    Status: StatusEnumActive,
    UserProfile: UserProfile{
        UserRole: &userRole, // *UserRoleEnum
    },
}

// Validate single object
if err := output.Validate(); err != nil {
    log.Fatal("Validation failed:", err)
}
```

### Multiple Object Validation

```go
import "github.com/your-org/dotprompt-gen-go/pkg/validator"

// Validate multiple outputs at once
if err := validator.ValidateAll(output1, output2, output3); err != nil {
    log.Fatal("Validation failed:", err)
}
```

### Custom Field Validation

```go
// Use the utility for custom validation scenarios
fieldValidations := map[string]validator.Validator{
    "custom_field": myCustomValidatableObject,
    "another_field": anotherValidatableObject,
}

if err := validator.ValidateFields(fieldValidations); err != nil {
    log.Fatal("Field validation failed:", err)
}
```

## Validation Rules

1. **Output structs only** - Input structs don't get master validation
2. **Enum fields** - Automatically validated against allowed values
3. **Object fields** - Nested structs with their own validation methods
4. **Pointer fields** - Safely handled with nil checks
5. **Required fields** - Validated according to schema requirements

## Benefits

- **Type-safe validation** at compile time
- **Comprehensive error messages** with field context
- **Hierarchical validation** - validates nested structures recursively
- **Zero configuration** - works automatically with generated code
- **Performance optimized** - early termination on first error
