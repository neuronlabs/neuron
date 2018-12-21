package mapping

import (
	"github.com/kucjac/jsonapi/pkg/internal/models"
)

type NestedStruct struct {
	*models.NestedStruct
}

// NestedField is the field within the NestedStruct
type NestedField struct {
	*models.NestedField
}
