package db

import (
	"github.com/neuronlabs/neuron/mapping"
)

func fieldSetWithUpdatedAt(model *mapping.ModelStruct, fields ...*mapping.StructField) mapping.FieldSet {
	updatedAt, hasUpdatedAt := model.UpdatedAt()
	if hasUpdatedAt {
		fields = append(fields, updatedAt)
	}
	return fields
}
