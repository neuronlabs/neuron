package mapping

import (
	"github.com/neuronlabs/strcase"
)

// Namer is the function that change the name with some prepared formatting.
type Namer func(string) string

// NamingSnake is a Namer function that converts the 'TestingModelName' into the 'testing_model_name' format.
func NamingSnake(raw string) string {
	return strcase.ToSnake(raw)
}

// NamingKebab is a Namer function that converts the 'TestingModelName' into the 'testing-model-name' format.
func NamingKebab(raw string) string {
	return strcase.ToKebab(raw)
}

// NamingCamel is a Namer function that converts the 'TestingModelName' into the 'TestingModelName' format.
func NamingCamel(raw string) string {
	return strcase.ToCamel(raw)
}

// NamingLowerCamel is a Namer function that converts the 'TestingModelName' into the 'testingModelName' format.
func NamingLowerCamel(raw string) string {
	return strcase.ToLowerCamel(raw)
}
