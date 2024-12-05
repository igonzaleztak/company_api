package binding

import (
	"strings"
	errors "xm_test/internal/api_errors"
	"xm_test/internal/enum"

	"github.com/go-playground/validator/v10"
)

func handleBindingErrors(err error) error {
	validationErrors := err.(validator.ValidationErrors)
	validationErr := validationErrors[0]

	fieldName := validationErr.Field()

	apiError := errors.ErrInvalidBody

	switch validationErr.Tag() {
	case "required":
		apiError.Message = fieldName + " is required and must be a " + validationErr.Type().String()
	case "oneof":
		apiError.Message = fieldName + " must be one of: " + strings.Join(strings.Split(validationErr.Param(), " "), ", ")
	case "customOneOf":
		apiError.Message = fieldName + " must be one of: " + strings.Join(enum.AllCompanyTypesString(), ", ")
	case "email":
		apiError.Message = fieldName + " must be a valid email address"
	default:
		apiError.Message = err.Error()
	}

	return apiError
}
