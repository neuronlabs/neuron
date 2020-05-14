package service

// namingConvention is the model mapping naming convention.
type namingConvention int

const (
	_ namingConvention = iota
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

func (n namingConvention) String() string {
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
