package filters

import (
	"fmt"
	aerrors "github.com/kucjac/jsonapi/errors"
	"github.com/kucjac/jsonapi/i18n"
	"github.com/kucjac/jsonapi/internal"
	"github.com/kucjac/jsonapi/internal/models"
	"golang.org/x/text/language"
	"reflect"
	"strings"
)

// FilterField is a field that contains information about filters
type FilterField struct {
	structField *models.StructField

	// Key is used for the map type filters as a nested argument
	key string

	// AttrFilters are the filter values for given attribute FilterField
	values []*OpValuePair

	// if given filterField is a relationship type it should be filter by it's
	// subfields (for given relation type).
	// Relationships are the filter values for given relationship FilterField
	nested []*FilterField

	raw string
}

// ToRaw returns raw string filter
func (f *FilterField) Raw() string {
	return ""
}

// Key returns the filter key if set
func (f *FilterField) Key() string {
	return f.key
}

// AddValues adds provided values to the filter values
func (f *FilterField) AddValues(values ...*OpValuePair) {
	f.values = append(f.values, values...)
}

// StructField returns the filter's StructField
func (f *FilterField) StructField() *models.StructField {
	return f.structField
}

// NestedFields returns nested fields for given filter
func (f *FilterField) NestedFields() []*FilterField {
	return f.nested
}

// Values return filter values
func (f *FilterField) Values() []*OpValuePair {
	return f.values
}

// NewFilter creates the filterField for given stuct field and values
func NewFilter(s *models.StructField, values ...*OpValuePair) *FilterField {
	return &FilterField{structField: s, values: values}
}

// FilterAppendvalues adds the values to the given filter field
func FilterAppendValues(f *FilterField, values ...*OpValuePair) {
	f.values = append(f.values, values...)
}

// FilterOpValuePairs returns the values of the provided filter field.
func FilterOpValuePairs(f *FilterField) []*OpValuePair {
	return f.values
}

// CopyFilter copies given FilterField with it's values
func CopyFilter(f *FilterField) *FilterField {
	return f.copy()
}

// GetOrCreateNestedFilter gets the filter field for given field
// If the field is a key returns the filter by key
func GetOrCreateNestedFilter(f *FilterField, field interface{}) *FilterField {
	return f.getOrCreateNested(field)
}

func NestedFields(f *FilterField) []*FilterField {
	return f.nested
}

// AddNestedField adds the nested field value
func (f *FilterField) AddNestedField(nested *FilterField) {
	addNestedField(f, nested)
}

func AddNestedField(f, nested *FilterField) {
	addNestedField(f, nested)
}

// AddsNestedField for given FilterField
func addNestedField(f, nested *FilterField) {

	for _, nf := range f.nested {
		if nf.structField == nested.structField {
			// Append the values to the given filter
			nf.values = append(nf.values, nested.values...)
			return
		}
	}

	f.nested = append(f.nested, nested)
	return

}

// SetValues sets the filter values for provided field, it's operator and possible i18n Support
func (f *FilterField) SetValues(
	values []string,
	op *Operator,
	sup *i18n.Support,
) (errObj *aerrors.ApiError) {
	var (
		er error

		opInvalid = func() {
			errObj = aerrors.ErrUnsupportedQueryParameter.Copy()
			errObj.Detail = fmt.Sprintf("The filter operator: '%s' is not supported for the field: '%s'.", op, f.structField.ApiName())
		}
	)

	t := f.structField.GetDereferencedType()
	// create new FilterValue
	fv := new(OpValuePair)

	fv.operator = op

	// Add and check all values for given field type
	switch f.structField.FieldKind() {
	case models.KindPrimary:
		switch t.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if op.Id > OpLessEqual.Id {
				opInvalid()
				return
			}
		case reflect.String:
			if !op.isBasic() {
				opInvalid()
				return
			}
		}

		for _, value := range values {

			fieldValue := reflect.New(t).Elem()
			er = setPrimaryField(value, fieldValue)
			if er != nil {
				errObj = aerrors.ErrInvalidQueryParameter.Copy()
				errObj.Detail = fmt.Sprintf("Invalid filter value for primary field in collection: '%s'. %s. ", f.structField.Struct().Collection(), er)
				return
			}
			fv.Values = append(fv.Values, fieldValue.Interface())
		}
		f.values = append(f.values, fv)

		// if it is of integer type check which kind of it
	case models.KindAttribute:
		switch t.Kind() {
		case reflect.String:
		default:
			if op.isStringOnly() {
				opInvalid()
				return
			}
		}
		if f.structField.IsLanguage() {
			switch op {
			case OpIn, OpEqual, OpNotIn, OpNotEqual:
				if sup != nil {
					for i, value := range values {
						tag, err := language.Parse(value)
						if err != nil {
							switch v := err.(type) {
							case language.ValueError:
								errObj = aerrors.ErrLanguageNotAcceptable.Copy()
								errObj.Detail = fmt.Sprintf("The value: '%s' for the '%s' filter field within the collection '%s', is not a valid language. Cannot recognize subfield: '%s'.", value, f.structField.ApiName(),
									f.structField.Struct().Collection(), v.Subtag())
								return
							default:
								errObj = aerrors.ErrInvalidQueryParameter.Copy()
								errObj.Detail = fmt.Sprintf("The value: '%v' for the '%s' filter field within the collection '%s' is not syntetatically valid.", value, f.structField.ApiName(), f.structField.Struct().Collection())
								return
							}
						}
						if op == OpEqual {
							var confidence language.Confidence
							tag, _, confidence = sup.Matcher.Match(tag)
							if confidence <= language.Low {
								errObj = aerrors.ErrLanguageNotAcceptable.Copy()
								errObj.Detail = fmt.Sprintf("The value: '%s' for the '%s' filter field within the collection '%s' does not match any supported langauges. The server supports following langauges: %s", value, f.structField.ApiName(), f.structField.Struct().Collection(),
									strings.Join(sup.PrettyLanguages(), internal.AnnotationSeperator),
								)
								return
							}
						}
						b, _ := tag.Base()
						values[i] = b.String()
					}
				}
			default:
				errObj = aerrors.ErrInvalidQueryParameter.Copy()
				errObj.Detail = fmt.Sprintf("Provided operator: '%s' for the language field is not acceptable", op.String())
				return
			}
		}
		for _, value := range values {
			fieldValue := reflect.New(t).Elem()
			er = setAttributeField(value, fieldValue)
			if er != nil {
				errObj = aerrors.ErrInvalidQueryParameter.Copy()
				errObj.Detail = fmt.Sprintf("Invalid filter value for the attribute field: '%s' for collection: '%s'. %s.", f.structField.ApiName(), f.structField.Struct().Collection(), er)
				return
			}
			fv.Values = append(fv.Values, fieldValue.Interface())
		}

		f.values = append(f.values, fv)

	}
	return
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

	default:
		return nil
		// invalid field type
	}

	return nil
}
