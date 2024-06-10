package pkg

import (
	"net/http"
	"reflect"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
)

var validate = validator.New()

type ValidationErrs map[string][]string

func ValidateRequest[T any](request T) (ValidationErrs, error) {
	errs := validate.Struct(request)
	validationErrs := ValidationErrs{}
	if errs != nil {
		// nolint:errorlint
		for _, err := range errs.(validator.ValidationErrors) {
			field, _ := reflect.TypeOf(request).Elem().FieldByName(err.Field())
			queryTag := getStructTag(field, "query")
			message := err.Tag() + ":" + getStructTag(field, "message")
			validationErrs[queryTag] = append(validationErrs[queryTag], message)
		}
		return validationErrs, errs
	}
	return nil, nil
}

// getStructTag returns the value of the tag with the given name
func getStructTag(f reflect.StructField, tagName string) string {
	return string(f.Tag.Get(tagName))
}

func BindRequest[T any](c echo.Context, request T) error {
	err := c.Bind(request)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return nil
}
