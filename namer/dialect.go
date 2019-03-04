package namer

import (
	"github.com/kucjac/jsonapi/mapping"
)

type DialectFieldNamer func(*mapping.StructField) string
