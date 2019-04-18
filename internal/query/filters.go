package query

import (
	"fmt"
	aerrors "github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/internal"
	"github.com/neuronlabs/neuron/internal/models"
	"github.com/neuronlabs/neuron/internal/query/filters"
	"github.com/neuronlabs/neuron/internal/query/scope"
	"strings"
)

// BuildRawFilter build  the filter from the provided values, with provided string values
func (b *Builder) BuildRawFilter(
	s *scope.Scope,
	raw string,
	values ...interface{},
) (fField *filters.FilterField, err error) {
	if len(values) == 1 {
		stringValues, ok := values[0].(string)
		if ok {
			return b.buildFilterField(s, raw, stringValues)
		}
	}

	return b.buildFilterWithValues(s, raw, values...)
}

func (b *Builder) buildFilterWithValues(
	s *scope.Scope,
	raw string,
	values ...interface{},
) (fField *filters.FilterField, err error) {
	/**

	TO IMPLEMENT

	*/
	return
}

func (b *Builder) buildFilterField(
	s *scope.Scope,
	raw string,
	rawValues string,
) (fField *filters.FilterField, err error) {

	var (
		sField *models.StructField
		op     *filters.Operator

		fieldName string

		// private function for returning ErrObject
		invalidName = func(fieldName, collection string) {
			errObj := aerrors.ErrInvalidQueryParameter.Copy()
			errObj.Detail = fmt.Sprintf("Provided invalid filter field name: '%s' for the '%s' collection.", fieldName, collection)
			err = errObj
			return
		}
		invalidOperator = func(operator string) {
			errObj := aerrors.ErrInvalidQueryParameter.Copy()
			errObj.Detail = fmt.Sprintf("Provided invalid filter operator: '%s' for the '%s' field.", operator, fieldName)
			err = errObj
			return
		}
	)

	// get other operators
	splitted, er := internal.SplitBracketParameter(raw[len(internal.QueryParamFilter):])
	if er != nil {
		errObj := aerrors.ErrInvalidQueryParameter.Copy()
		errObj.Detail = fmt.Sprintf("The filter paramater is of invalid form. %s", er)
		err = errObj
		return
	}

	schema, ok := b.schemas.Schema(s.Struct().SchemaName())
	if !ok {
		err = internal.IErrModelNotMapped
		return
	}

	collection := splitted[0]

	splitted = splitted[1:]

	colModel := schema.ModelByCollection(collection)
	if colModel == nil {
		errObj := aerrors.ErrInvalidQueryParameter.Copy()
		errObj.Detail = fmt.Sprintf("Provided invalid collection: '%s' in the filter query.", collection)
		err = errObj
		return
	}
	m := colModel

	filterScope := s.GetModelsRootScope(colModel)
	if filterScope == nil {
		errObj := aerrors.ErrInvalidQueryParameter.Copy()
		errObj.Detail = fmt.Sprintf("The collection: '%s' is not included in query.", collection)
		err = errObj
		return
	}

	values := strings.Split(rawValues, internal.AnnotationSeperator)

	// check if any parameters are set for filtering
	if len(splitted) == 0 {
		errObj := aerrors.ErrInvalidQueryParameter.Copy()
		errObj.Detail = fmt.Sprint("Too few filter parameters. Valid format is: filter[collection][field][subfield|operator]([operator])*.")
		err = errObj
		return
	}

	// for all cases first value should be a fieldName
	fieldName = splitted[0]

	if len(values) > b.Config.FilterValueLimit {
		errObj := aerrors.ErrOutOfRangeQueryParameterValue.Copy()
		errObj.Detail = fmt.Sprintf("The number of values for the filter: %s within collection: %s exceeds the permissible length: '%d'", strings.Join(splitted, internal.AnnotationSeperator), collection, b.Config.FilterValueLimit)
		err = errObj
		return
	}

	switch len(splitted) {
	case 1:
		// if there is only one argument it must be an attribute.
		if fieldName == "id" {
			fField = s.GetOrCreateIDFilter()
		} else {
			sField, ok = m.Attribute(fieldName)
			if !ok {
				_, ok = m.RelationshipField(fieldName)
				if ok {
					errObj := aerrors.ErrInvalidQueryParameter.Copy()
					errObj.Detail = fmt.Sprintf("Provided filter field: '%s' is a relationship. In order to filter a relationship specify the relationship's field. i.e. '/%s?filter[%s][id]=123'", fieldName, m.Collection(), fieldName)
					err = errObj
					return
				}

				sField, ok = m.FilterKey(fieldName)
				if !ok {
					invalidName(fieldName, m.Collection())
					return
				}
				fField = s.GetOrCreateFilterKeyFilter(sField)
			} else {
				if sField.IsLanguage() {
					fField = s.GetOrCreateLanguageFilter()
				} else if sField.IsMap() {
					// Map doesn't allow any default parameters
					// i.e. filter[collection][mapfield]=something
					errObj := aerrors.ErrInvalidQueryParameter.Copy()
					errObj.Detail = "Cannot filter field of 'Hashmap' type with no operator"
					err = errObj
					return
				} else {
					fField = s.GetOrCreateAttributeFilter(sField)
				}
			}
		}

		errObj := fField.SetValues(values, filters.OpEqual, b.I18n)
		if errObj != nil {
			err = errObj
			return
		}
	case 2:
		if fieldName == "id" {
			fField = s.GetOrCreateIDFilter()
		} else {
			sField, ok = m.Attribute(fieldName)
			if !ok {
				// jeżeli relacja ->
				sField, ok = m.RelationshipField(fieldName)
				if !ok {
					invalidName(fieldName, m.Collection())
					return
				}

				// if field were already used
				fField = s.GetOrCreateRelationshipFilter(sField)

				err = b.buildNestedFilter(fField, values, splitted[1:]...)

				return
			}
			if sField.IsLanguage() {
				fField = s.GetOrCreateLanguageFilter()
			} else {
				fField = s.GetOrCreateAttributeFilter(sField)
			}
		}
		// it is an attribute filter
		op, ok = filters.Operators.Get(splitted[1])
		if !ok {
			invalidOperator(splitted[1])
			return
		}

		errObj := fField.SetValues(values, op, b.I18n)
		if errObj != nil {
			err = errObj
			return
		}

	case 3:
		// musi być relacja
		sField, ok = m.RelationshipField(fieldName)
		if !ok {
			// moze ktos podal attribute
			_, ok = m.Attribute(fieldName)
			if !ok {
				invalidName(fieldName, m.Collection())
				return
			}
			errObj := aerrors.ErrInvalidQueryParameter.Copy()
			errObj.Detail = fmt.Sprintf("Too many parameters for the attribute field: '%s'.", fieldName)
			err = errObj
			return
		}
		fField = s.GetOrCreateRelationshipFilter(sField)

		errObj := b.buildNestedFilter(fField, values, splitted[1:]...)
		if errObj != nil {
			err = errObj
		}

	default:
		var errObj *aerrors.ApiError

		_, ok = m.Attribute(fieldName)
		if !ok {
			_, ok = m.RelationshipField(fieldName)
			if !ok {
				errObj = aerrors.ErrInvalidQueryParameter.Copy()
				errObj.Detail = fmt.
					Sprintf("Invalid field name: '%s' for '%s' collection.", fieldName, collection)
				err = errObj
			}
		}
		if errObj == nil {
			errObj = aerrors.ErrInvalidQueryParameter.Copy()
			errObj.Detail = fmt.
				Sprintf("Too many filter parameters for '%s' collection. ", collection)
			err = errObj
		}
	}
	return

}

func (b *Builder) buildNestedFilter(
	f *filters.FilterField,
	values []string,
	splitted ...string,
) (errObj *aerrors.ApiError) {

	// internal variables definitions
	var (
		subfield *models.StructField
		op       *filters.Operator
		ok       bool
	)

	//	internal function definitions
	var (
		addDetails = func() {
			errObj.Detail += fmt.Sprintf("Collection: '%s', FilterField: '%s'", f.StructField().Struct().Collection(), f.StructField().ApiName())
		}

		getSubfield = func() (subfield *models.StructField, eObj *aerrors.ApiError) {
			var ok bool
			if splitted[0] == "id" {
				subfield = models.FieldsRelatedModelStruct(f.StructField()).PrimaryField()
				return
			}
			subfield, ok = models.FieldsRelatedModelStruct(f.StructField()).Attribute(splitted[0])
			if !ok {
				eObj = aerrors.ErrInvalidQueryParameter.Copy()
				eObj.Detail = fmt.Sprintf("The subfield name: '%s' is invalid.", splitted[0])
				return
			}
			return
		}
	)

	// if the field is of relationship kind
	if f.StructField().IsRelationship() {
		// Get the subfield
		subfield, errObj = getSubfield()
		if errObj != nil {
			addDetails()
			return
		}

		filter := filters.GetOrCreateNestedFilter(f, subfield)

		// Check if the subfield is a map type
		if subfield.IsMap() {

			// the splitted[0] is hash map and there is no operator
			if len(splitted) == 1 {

				// if no operators provided the value is an error

				errObj = aerrors.ErrInvalidQueryParameter.Copy()
				errObj.Detail = fmt.Sprintf("The filter subfield: '%s' is of HashMap type. This field type requires at least 'key' filter.", splitted[0])
				addDetails()
				return errObj

				// the splitted[0] is map, check the splitted[1] if it is an operator or key
			} else {

				// otherwise check if the filter contains map->key or the operator
				_, ok := filters.Operators.Get(splitted[1])
				if ok {

					// filter[colleciton][relationship][map][$eq]

					// the filter to the map cannot be compared with an operator (with no key)
					errObj = aerrors.ErrInvalidQueryParameter.Copy()
					errObj.Detail = fmt.Sprint("The HashMap type field cannot be comparable using raw operators. Add 'key' filter to the query.")
					addDetails()
					return
				}

				// otherwise the value must be a map - key
				// filter[collection][relationship][map][attr]
				if len(values) > 1 {
					op = filters.OpIn
				} else {
					op = filters.OpEqual
				}

				interfaceValues := []interface{}{}

				for _, v := range values {
					interfaceValues = append(interfaceValues, v)
				}

				// Get the Subfield Filtervalue
				fv := filters.NewOpValuePair(op, interfaceValues...)

				// getOrCreateNested field for given 'key'
				nested := filters.GetOrCreateNestedFilter(filter, splitted[1])

				// add the 'key' values
				nested.AddValues(fv)
			}

			// if the relatinoship subfield is not a map - create just raw relationship subfield
		} else {

			// when no operator is provided set based on the value length
			if len(splitted) == 1 {
				if len(values) > 1 {
					op = filters.OpIn
				} else {
					op = filters.OpEqual
				}

				// otherwise get it from the splitted[1]
			} else {
				operator := splitted[1]
				op, ok = filters.Operators.Get(operator)
				if !ok {
					// invalid operator is provided or nested structure.
					errObj = aerrors.ErrInvalidQueryParameter.Copy()
					errObj.Detail = fmt.Sprintf("Provided invalid filter operator: '%s' on collection:'%s'.", operator, f.StructField().Struct().Collection())
					return errObj
				}
			}

			// set filter values for the given filter field
			errObj = filter.SetValues(values, op, b.I18n)

			// no need to add subfield filter, it is already set by the function getOrCreateNested
		}

		// if the filter field is a map type attribute, check the operators or keys
	} else if f.StructField().IsMap() {

		// if the filter contains single parameter over the map
		if len(splitted) == 1 {
			// otherwise check if the filter contains map->key or the operator
			_, ok = filters.Operators.Get(splitted[0])
			if ok {

				// the filter to the map must be a
				errObj = aerrors.ErrInvalidQueryParameter.Copy()
				errObj.Detail = fmt.Sprint("The HashMap type field cannot be comparable using raw operators. Add 'key' filter to the query.")
				addDetails()
				return
			}

			// otherwise the value must be a map - key
			if len(values) > 1 {
				op = filters.OpIn
			} else {
				op = filters.OpEqual
			}
			key := splitted[0]

			// get or create (if already exists) nested filter field
			filter := filters.GetOrCreateNestedFilter(f, key)

			// when there are more filter paramters over the map

			interfaceVals := []interface{}{}
			for _, val := range values {
				interfaceVals = append(interfaceVals, val)
			}
			fv := filters.NewOpValuePair(op, interfaceVals...)
			filter.AddValues(fv)
			// the map filter should be done already

			// if the map filter contains the additional arguments they
			// should be of type: 'filter[collection][map][attr][$operator]'
		} else {

			// splitted[1] is a map -> key
			key := splitted[0]

			op, ok := filters.Operators.Get(splitted[1])
			if !ok {
				errObj := aerrors.ErrInvalidQueryParameter.Copy()
				errObj.Detail = fmt.Sprintf("Provided invalid filter operator: '%s' on collection:'%s, field: %v, and key: %s.'.", splitted[1], f.StructField().Struct().Collection(), f.StructField().ApiName(), key)
				return errObj
			}

			filter := filters.GetOrCreateNestedFilter(f, key)

			interfaceVals := []interface{}{}
			for _, v := range values {
				interfaceVals = append(interfaceVals, v)
			}

			fv := filters.NewOpValuePair(op, interfaceVals...)
			filter.AddValues(fv)
		}

		// no other filter field type is known
	} else {
		// error
		errObj = aerrors.ErrInvalidQueryParameter.Copy()
		errObj.Detail = fmt.Sprintf("The filter over field: %v contains too many arguments.", f.StructField().ApiName())
	}

	return
}
