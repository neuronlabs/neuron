package namer

import (
	"github.com/kucjac/jsonapi/mapping"
)

// DialectFieldNamer is the namer function
type DialectFieldNamer func(*mapping.StructField) string
