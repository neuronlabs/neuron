package namer

import (
	"github.com/kucjac/jsonapi/pkg/mapping"
)

type DialectFieldNamer func(*mapping.StructField) string
