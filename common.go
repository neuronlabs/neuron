package jsonapi

import (
	"github.com/iancoleman/strcase"
)

type Collectioner interface {
	CollectionName() string
}

func getNameByConvention(raw string, c NamingConvention) string {
	var name string
	switch c {
	case NamingSnake:
		name = strcase.ToSnake(raw)
	case NamingKebab:
		name = strcase.ToKebab(raw)
	case NamingCamel:
		name = strcase.ToCamel(raw)
	case NamingLowerCamel:
		name = strcase.ToLowerCamel(raw)
	}
	return name
}
