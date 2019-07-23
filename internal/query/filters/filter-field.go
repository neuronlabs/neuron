package filters

import (
	"reflect"

	"golang.org/x/text/language"

	"github.com/neuronlabs/neuron-core/errors"
	"github.com/neuronlabs/neuron-core/errors/class"
	"github.com/neuronlabs/neuron-core/i18n"

	"github.com/neuronlabs/neuron-core/internal/models"
)

// FilterField is a structure that contains filtering information about given field
// or it's subfields (in case of relationships).
type FilterField struct {
	structField *models.StructField

	// key is used for the map type filters as a nested argument.
	key string

	// values are the filter values for given attribute FilterField
	values []*OpValuePair

	// if given filterField is a relationship type it should be filter by it's
	// subfields (for given relation type).
	// Relationships are the filter values for given relationship FilterField
	nested []*FilterField
}

// Key returns the filter key if set.
func (f *FilterField) Key() string {
	return f.key
}

// AddNestedField adds the nested field value.
func (f *FilterField) AddNestedField(nested *FilterField) {
	addNestedField(f, nested)
}

// AddValues adds given operator value pairs to the filter values.
func (f *FilterField) AddValues(values ...*OpValuePair) {
	f.values = append(f.values, values...)
}

// AppendValues adds the values to the given filter field.
func (f *FilterField) AppendValues(values ...*OpValuePair) {
	f.values = append(f.values, values...)
}

// GetOrCreateNestedFilter gets the filter field for given field
// If the field is a key returns the filter by key
func (f *FilterField) GetOrCreateNestedFilter(field interface{}) *FilterField {
	return f.getOrCreateNested(field)
}

// NestedFields returns nested fields for given filter.
func (f *FilterField) NestedFields() []*FilterField {
	return f.nested
}

// StructField returns the filter's StructField.
func (f *FilterField) StructField() *models.StructField {
	return f.structField
}

// Values returns filter operator value pairs.
func (f *FilterField) Values() []*OpValuePair {
	return f.values
}

// CopyFilter deeply copies given FilterField with it's values.
func CopyFilter(f *FilterField) *FilterField {
	return f.copy()
}

// NewFilter creates the filterField for given models.StructField 's' and operator
// value pairs 'values'.
func NewFilter(s *models.StructField, values ...*OpValuePair) *FilterField {
	return &FilterField{structField: s, values: values}
}

// AddNestedField adds nested filterfield for the filter 'f'
func AddNestedField(f, nested *FilterField) {
	addNestedField(f, nested)
}

// AddsNestedField for given FilterField
func addNestedField(f, nested *FilterField) {
	// check if there already exists a nested filter
	for _, nf := range f.nested {

		if nf.structField == nested.structField {
			// Append the values to the given filter
			nf.values = append(nf.values, nested.values...)
			return
		}
	}

	f.nested = append(f.nested, nested)
}

// SetValues sets the filter values for provided field, it's operator and possible i18n Support.
func (f *FilterField) SetValues(values []string, op *Operator, sup *i18n.Support) error {
	t := f.structField.GetDereferencedType()

	// create new FilterValue
	fv := new(OpValuePair)
	fv.operator = op

	// Check all values for given field type
	switch f.structField.FieldKind() {
	case models.KindPrimary, models.KindForeignKey:
		switch t.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			// check if id is of proper type
			if op.isStringOnly() || op == OpIsNull || op == OpNotNull || op == OpNotExists || op == OpExists {
				// operations over
				err := errors.Newf(class.QueryFilterUnsupportedOperator, "filter operator is not supported: '%s' for given field", op)
				err = err.SetDetailf("The filter operator: '%s' is not supported for the field: '%s'.", op, f.structField.NeuronName())
				return err
			}
		case reflect.String:
			if !op.isBasic() {
				err := errors.Newf(class.QueryFilterUnsupportedOperator, "filter operator is not supported: '%s' for given field", op)
				err = err.SetDetailf("The filter operator: '%s' is not supported for the field: '%s'.", op, f.structField.NeuronName())
				return err
			}
		}

		for _, value := range values {
			fieldValue := reflect.New(t).Elem()

			err := setPrimaryField(value, fieldValue)
			if err != nil {
				if err.Class.IsMajor(class.MjrInternal) {
					return err
				}
				err = err.WrapDetailf("Invalid filter value for primary field in collection: '%s'.", f.structField.Struct().Collection())
				return err
			}
			fv.Values = append(fv.Values, fieldValue.Interface())
		}
		f.values = append(f.values, fv)

	case models.KindAttribute, models.KindFilterKey:
		// if it is of integer type check which kind of it
		switch t.Kind() {
		case reflect.String:
		default:
			// check if the operator is the string only - doesn't allow other types of values
			if op.isStringOnly() {
				err := errors.Newf(class.QueryFilterUnsupportedOperator, "string filter operator is not supported: '%s' for non string field", op)
				err = err.SetDetailf("The filter operator: '%s' is not supported for the field: '%s'.", op, f.structField.NeuronName())
				return err
			}
		}

		// if the field is a language
		if f.structField.IsLanguage() {
			switch op {
			case OpIn, OpEqual, OpNotIn, OpNotEqual:
				if sup != nil {
					for i, value := range values {
						tag, err := language.Parse(value)
						if err != nil {
							switch v := err.(type) {
							case language.ValueError:
								err := errors.New(class.QueryFilterLanguage, v.Error())
								err = err.SetDetailf("The value: '%s' for the '%s' filter field within the collection '%s', is not a valid language. Cannot recognize subfield: '%s'.", value, f.structField.NeuronName(),
									f.structField.Struct().Collection(), v.Subtag())
								return err
							default:
								return errors.New(class.QueryFilterLanguage, err.Error())
							}
						}
						if op == OpEqual {
							var confidence language.Confidence
							tag, _, confidence = sup.Matcher.Match(tag)
							if confidence <= language.Low {
								err := errors.New(class.QueryFilterLanguage, "unsupported language filter")
								err = err.SetDetailf("The value: '%s' for the '%s' filter field within the collection '%s' does not match any supported languages.", value, f.structField.NeuronName(), f.structField.Struct().Collection())
								return err
							}
						}
						b, _ := tag.Base()
						values[i] = b.String()
					}
				}
			default:
				err := errors.New(class.QueryFilterLanguage, "invalid query language filter operator")
				err = err.SetDetailf("Provided operator: '%s' for the language field is not acceptable", op.String())
				return err
			}
		}

		// iterate over filter values and check if they're correct
		for _, value := range values {
			fieldValue := reflect.New(t).Elem()
			err := setAttributeField(value, fieldValue)
			if err != nil {
				if err.Class.IsMajor(class.MjrInternal) {
					return err
				}

				err = err.WrapDetailf("Invalid filter value for the attribute field: '%s' for collection: '%s'.", f.structField.NeuronName(), f.structField.Struct().Collection())
				return err
			}
			fv.Values = append(fv.Values, fieldValue.Interface())
		}

		f.values = append(f.values, fv)
	default:
		return errors.Newf(class.InternalQueryFilter, "setting filter values for invalid field: '%s'", f.structField.Name())
	}
	return nil
}

func (f *FilterField) copy() *FilterField {
	dst := &FilterField{structField: f.structField}

	if len(f.values) != 0 {
		dst.values = make([]*OpValuePair, len(f.values))
		for i, value := range f.values {
			dst.values[i] = value.copy()
		}
	}

	if len(f.nested) != 0 {
		dst.nested = make([]*FilterField, len(f.nested))
		for i, value := range f.nested {
			dst.nested[i] = value.copy()
		}
	}
	return dst
}

// getNested gets the nested field by it's key or structfield
func (f *FilterField) getOrCreateNested(field interface{}) *FilterField {

	switch t := field.(type) {
	case string:
		for _, nested := range f.nested {
			if nested.key == t {
				return nested
			}
		}
		nested := &FilterField{key: t}
		f.nested = append(f.nested, nested)
		return nested
	case *models.StructField:
		for _, nested := range f.nested {
			if nested.structField == t {
				return nested
			}
		}

		nested := &FilterField{structField: t}
		f.nested = append(f.nested, nested)
		return nested
	}
	return nil
}
