package jsonapi

type NamingConvention int

const (
	NamingSnake      NamingConvention = iota //any_kind_of_string
	NamingKebab                              // any-kind-of-string
	NamingCamel                              // AnyKindOfString
	NamingLowerCamel                         // anyKindOfString
)
