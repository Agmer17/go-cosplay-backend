package utils

import (
	"errors"
	"strings"

	"github.com/go-playground/validator/v10"
)

func ParseValidationErrors(err error) (map[string]string, bool) {
	var ve validator.ValidationErrors

	if !errors.As(err, &ve) {
		return nil, false
	}

	out := make(map[string]string)

	for _, fe := range ve {
		field := strings.ToLower(fe.Field())
		out[field] = fe.Error()
	}

	return out, true
}
