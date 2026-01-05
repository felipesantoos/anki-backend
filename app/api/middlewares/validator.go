package middlewares

import (
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// CustomValidator is a custom validator that implements echo.Validator interface
type CustomValidator struct {
	validator *validator.Validate
}

// NewCustomValidator creates a new CustomValidator instance
func NewCustomValidator() *CustomValidator {
	return &CustomValidator{
		validator: validator.New(),
	}
}

// Validate validates the struct using go-playground/validator
func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		// Return validation errors in a user-friendly format
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			return echo.NewHTTPError(400, formatValidationErrors(validationErrors))
		}
		return echo.NewHTTPError(400, err.Error())
	}
	return nil
}

// formatValidationErrors formats validation errors into a readable message
func formatValidationErrors(errs validator.ValidationErrors) string {
	if len(errs) == 0 {
		return "Validation failed"
	}
	
	// Return the first error message for simplicity
	// In production, you might want to return all errors
	err := errs[0]
	
	switch err.Tag() {
	case "required":
		return err.Field() + " is required"
	case "gt":
		return err.Field() + " must be greater than " + err.Param()
	case "gte":
		return err.Field() + " must be greater than or equal to " + err.Param()
	case "lt":
		return err.Field() + " must be less than " + err.Param()
	case "lte":
		return err.Field() + " must be less than or equal to " + err.Param()
	case "min":
		return err.Field() + " must be at least " + err.Param()
	case "max":
		return err.Field() + " must be at most " + err.Param()
	case "email":
		return err.Field() + " must be a valid email address"
	case "oneof":
		return err.Field() + " must be one of: " + err.Param()
	default:
		return err.Field() + " is invalid"
	}
}

