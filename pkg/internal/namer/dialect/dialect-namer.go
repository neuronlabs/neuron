package dialect

import (
	"github.com/kucjac/jsonapi/pkg/internal/models"
)

type DialectFieldNamer func(*models.StructField) string
