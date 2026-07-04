package validation

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/shivam/featfz/feat-manager/internal/http/response"
)

func NewValidator() *validator.Validate {
	validate := validator.New(validator.WithRequiredStructEnabled())
	validate.RegisterTagNameFunc(func(field reflect.StructField) string {
		tag := field.Tag.Get("json")
		name := strings.Split(tag, ",")[0]
		if name == "" || name == "-" {
			return field.Name
		}

		return name
	})

	return validate
}

func ValidationDetails(err error) []response.ErrorDetail {
	if err == nil {
		return nil
	}

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return []response.ErrorDetail{{Field: "body", Message: "contains invalid data"}}
	}

	details := make([]response.ErrorDetail, 0, len(validationErrors))
	for _, validationError := range validationErrors {
		details = append(details, response.ErrorDetail{
			Field:   validationError.Field(),
			Message: validationMessage(validationError),
		})
	}

	return details
}

func validationMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "is required"
	case "max":
		return fmt.Sprintf("must be at most %s characters", err.Param())
	case "min":
		return fmt.Sprintf("must be at least %s characters", err.Param())
	default:
		return fmt.Sprintf("failed validation: %s", err.Tag())
	}
}
