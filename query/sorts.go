package query

import (
	"fmt"
	"github.com/neuronlabs/neuron-core/internal/models"
	"net/url"

	"github.com/neuronlabs/neuron-core/mapping"

	"github.com/neuronlabs/neuron-core/internal/query/sorts"
)

// ParamSort is the url query parameter name for the sorting fields.
const ParamSort = "sort"

// SortField is a field that contains sorting information.
type SortField sorts.SortField

// StructField returns sortfield's structure.
func (s *SortField) StructField() *mapping.StructField {
	sField := (*sorts.SortField)(s).StructField()

	return (*mapping.StructField)(sField)
}

// Order returns sortfield's order.
func (s *SortField) Order() SortOrder {
	return SortOrder((*sorts.SortField)(s).Order())
}

// FormatQuery returns the sort field formatted for url query.
// If the optional argument 'q' is provided the format would be set into the provdied url.Values.
// Otherwise it creates new url.Values instance.
// Returns modified url.Values
func (s *SortField) FormatQuery(q ...url.Values) url.Values {
	var query url.Values
	if len(q) > 0 {
		query = q[0]
	}

	if query == nil {
		query = url.Values{}
	}

	var sign string
	if s.Order() == DescendingOrder {
		sign = "-"
	}

	var v string
	if vals, ok := query[ParamSort]; ok {
		if len(vals) > 0 {
			v = vals[0]
		}

		if len(v) > 0 {
			v += ","
		}
	}

	v += fmt.Sprintf("%s%s", sign, s.StructField().NeuronName())

	query.Set(ParamSort, v)

	return query
}

// SortOrder is the enum used as the sorting values order.
type SortOrder int

const (
	// AscendingOrder defines the sorting ascending order.
	AscendingOrder SortOrder = iota

	// DescendingOrder defines the sorting descending order.
	DescendingOrder
)

// String implements fmt.Stringer interface.
func (o SortOrder) String() string {
	if o == AscendingOrder {
		return "ascending"
	}
	return "descending"
}

// NewSortFields creates new 'sortFields' for given model 'm'. If the 'disallowFK' is set to true
// the function would not allow to create foreign key sort field.
// The function throws errors on duplicated field values.
func NewSortFields(m *mapping.ModelStruct, disallowFK bool, sortFields ...string) ([]*SortField, error) {
	sortFieldStructs, err := sorts.NewUniques((*models.ModelStruct)(m), disallowFK, sortFields...)
	if err != nil {
		return nil, err
	}

	thisFields := make([]*SortField, len(sortFieldStructs))
	for i, sField := range sortFieldStructs {
		thisFields[i] = (*SortField)(sField)
	}
	return thisFields, nil
}

// NewSort creates new 'sort' field for given model 'm'. If the 'disallowFK' is set to true
// the function would not allow to create Sort field of foreign key field.
func NewSort(m *mapping.ModelStruct, sort string, disallowFK bool) (*SortField, error) {
	sField, err := sorts.New((*models.ModelStruct)(m), sort, disallowFK)
	if err != nil {
		return nil, err
	}
	return (*SortField)(sField), nil
}
