package grok

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Error ...
type Error struct {
	Code     int      `json:"code"`
	Key      string   `json:"key"`
	Messages []string `json:"messages"`
}

// NewError  ...
func NewError(code int, key string, messages ...string) *Error {
	return &Error{Code: code, Key: key, Messages: messages}
}

// FromValidationErros ...
func FromValidationErros(errors error) *Error {
	validationErrors, ok := errors.(validator.ValidationErrors)

	if !ok {
		return NewError(0, "cannot parse validation errors")
	}

	err := NewError(http.StatusUnprocessableEntity, "INVALID_PARAMETER")

	message := "validation failed for %s"

	for _, e := range validationErrors {
		err.Messages = append(err.Messages, fmt.Sprintf(message, e.Field()))
	}

	return err
}

func (e *Error) Error() string {
	return fmt.Sprintf(
		"Code: %d - Messages: %s",
		e.Code,
		strings.Join(e.Messages, "\n"),
	)
}
