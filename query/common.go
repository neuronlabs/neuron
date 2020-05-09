package query

import (
	"github.com/neuronlabs/errors"

	"github.com/neuronlabs/neuron/class"
	"github.com/neuronlabs/neuron/mapping"
)

func fieldSetWithUpdatedAt(model *mapping.ModelStruct, fields ...*mapping.StructField) mapping.FieldSet {
	updatedAt, hasUpdatedAt := model.UpdatedAt()
	if hasUpdatedAt {
		fields = append(fields, updatedAt)
	}
	return fields
}

func (s *Scope) requireNoFilters() error {
	if len(s.Filters) != 0 {
		return errors.Newf(class.QueryNotValid, "given query doesn't allow filtering")
	}
	return nil
}

func (s *Scope) logFormat(format string) string {
	return "SCOPE[" + s.id.String() + "]" + format
}
