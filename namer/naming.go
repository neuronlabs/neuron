package namer

import (
	"github.com/iancoleman/strcase"
)

// Namer is the function that change the name with some prepared formatting.
type Namer func(string) string

// NamingSnake is a Namer function that converts the 'raw' into the 'snake_case_model'.
func NamingSnake(raw string) string {
	return strcase.ToSnake(raw)
}

// NamingKebab is a Namer function that converts the 'raw' into the 'kebab-case-model'.
func NamingKebab(raw string) string {
	return strcase.ToKebab(raw)
}

// NamingCamel is a Namer function that converts the 'raw' into the 'CamelCaseModel'.
func NamingCamel(raw string) string {
	return strcase.ToCamel(raw)
}

// NamingLowerCamel is a Namer function that converts the 'raw' into the 'camelCaseModel'.
func NamingLowerCamel(raw string) string {
	return strcase.ToLowerCamel(raw)
}
