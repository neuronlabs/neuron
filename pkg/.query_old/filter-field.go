package query

import (
	"github.com/kucjac/jsonapi/pkg/internal/models"
	"github.com/kucjac/jsonapi/pkg/mapping"
)

// FilterField is a field that contains information about filters
type FilterField struct {
	*mapping.StructField

	// Key is used for the map type filters as a nested argument
	Key string

	// AttrFilters are the filter values for given attribute FilterField
	Values []*FilterValues

	// if given filterField is a relationship type it should be filter by it's
	// subfields (for given relation type).
	// Relationships are the filter values for given relationship FilterField
	Nested []*FilterField
}

func (f *FilterField) copy() *FilterField {
	dst := &FilterField{StructField: f.StructField}

	if len(f.Values) != 0 {
		dst.Values = make([]*FilterValues, len(f.Values))
		for i, value := range f.Values {
			dst.Values[i] = value.copy()
		}
	}

	if len(f.Nested) != 0 {
		dst.Nested = make([]*FilterField, len(f.Nested))
		for i, value := range f.Nested {
			dst.Nested[i] = value.copy()
		}
	}
	return dst
}

// getNested gets the nested field by it's key or structfield
func (f *FilterField) getOrCreateNested(field interface{}) *FilterField {
	if models.FieldIsMap(f.StructField.StructField) {
		key := field.(string)
		for _, nested := range f.Nested {
			if nested.Key == key {
				return nested
			}
		}
		nested := &FilterField{Key: key}
		f.Nested = append(f.Nested, nested)
		return nested
	} else {
		sField := field.(*mapping.StructField)
		for _, nested := range f.Nested {
			if nested.StructField == sField {
				return nested
			}
		}
		nested := &FilterField{StructField: sField}
		f.Nested = append(f.Nested, nested)
		return nested
	}
	return nil
}
