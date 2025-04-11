package validator

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// CustomValidator wraps the validator instance
type CustomValidator struct {
	validate *validator.Validate
}

// NewValidator initializes and returns a new CustomValidator
func NewValidator() *CustomValidator {
	return &CustomValidator{
		validate: validator.New(),
	}
}

// Validate validates any struct based on validation tags
func (cv *CustomValidator) Validate(i interface{}) error {
	err := cv.validate.Struct(i)
	if err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			first := validationErrors[0]
			field := strings.ToLower(first.Field())
			tag := first.Tag()

			// Example error: "name" is required
			msg := fmt.Sprintf("\"%s\" is %s", field, tag)
			return errors.New(msg)
		}
		return errors.New("validation failed")
	}
	return nil
}
