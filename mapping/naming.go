package mapping

import (
	"strings"

	"github.com/neuronlabs/strcase"

	"github.com/neuronlabs/neuron/errors"
)

// namingConvention is the model mapping naming convention.
type NamingConvention int

const (
	_ NamingConvention = iota
	// SnakeCase is the naming convention where all words are in lower case letters separated by the '_' character.
	// i.e.: naming_convention
	SnakeCase
	// CamelCase is the naming convention where words are not separated by any character or space and each word starts
	// with a capital letter.
	// i.e.: NamingConvention
	CamelCase
	// LowerCamelCase is the naming convention where words are not separated by any character or space and all but first words starts
	// with a capital letter.
	// i.e.: namingConvention
	LowerCamelCase
	// KebabCase is the naming convention where all words are in lower case letters separated by the '-' character.
	// i.e.: naming-convention
	KebabCase
)

func (n *NamingConvention) Parse(name string) error {
	switch strings.ToLower(name) {
	case "snake":
		*n = SnakeCase
	case "lower_camel":
		*n = LowerCamelCase
	case "camel":
		*n = CamelCase
	case "kebab":
		*n = KebabCase
	default:
		return errors.WrapDetf(ErrNamingConvention, "unknown naming convention name: %s", name)
	}
	return nil
}

func (n NamingConvention) Namer(raw string) string {
	switch n {
	case SnakeCase:
		return NamingSnake(raw)
	case CamelCase:
		return NamingCamel(raw)
	case LowerCamelCase:
		return NamingLowerCamel(raw)
	case KebabCase:
		return NamingKebab(raw)
	default:
		return raw
	}
}

func (n NamingConvention) String() string {
	switch n {
	case SnakeCase:
		return "snake"
	case CamelCase:
		return "camel"
	case LowerCamelCase:
		return "lower_camel"
	case KebabCase:
		return "kebab"
	}
	return "unknown"
}

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
