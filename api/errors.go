package api

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

type ErrorResponse struct {
	Message string       `json:"message,omitempty"`
	Fields  []ErrorField `json:"errors,omitempty"`
}

type ValidationFormatter struct {
	structToJson map[string]string
}

type ErrorField struct {
	FieldName string `json:"field_name"`
	Message   string `json:"message"`
}

func (v ValidationFormatter) getErrorMsg(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required"
	case "lte":
		return fmt.Sprintf("Should be less than %s", fe.Param())
	case "gte":
		return fmt.Sprintf("Should be greater than %s", fe.Param())
	}
	return "Unknown error"
}

func (v ValidationFormatter) validate(err error) []ErrorField {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		errList := make([]ErrorField, len(ve))
		for i, fe := range ve {
			errList[i] = ErrorField{v.structToJson[fe.Field()], v.getErrorMsg(fe)}
		}
		return errList
	}

	return []ErrorField{}
}

func BuildValidationError(err error, v ValidationFormatter) ErrorResponse {
	var validationError validator.ValidationErrors
	if !errors.As(err, &validationError) {
		return ErrorResponse{
			Message: "Error isn't a validationError",
		}
	}

	return ErrorResponse{
		Fields: v.validate(err),
	}
}
