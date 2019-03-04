package dialect

import (
	"github.com/kucjac/jsonapi/internal/models"
)

type DialectFieldNamer func(*models.StructField) string
