package namer

import (
	"github.com/iancoleman/strcase"
)

// Namer is the interface responsible for naming the structures.
type Namer func(string) string

// NamingSnake is a Namer function.
// it convert the name into the 'snake_case_model'.
func NamingSnake(raw string) string {
	return strcase.ToSnake(raw)
}

// NamingKebab is a Namer function.
// it convert the name into the 'kebab-case-model'.
func NamingKebab(raw string) string {
	return strcase.ToKebab(raw)
}

// NamingCamel is a Namer function.
// it convert the name into the 'CamelCaseModel'.
func NamingCamel(raw string) string {
	return strcase.ToCamel(raw)
}

// NamingLowerCamel is a Namer function.
// it convert the name into the 'camelCaseModel'.
func NamingLowerCamel(raw string) string {
	return strcase.ToLowerCamel(raw)
}
