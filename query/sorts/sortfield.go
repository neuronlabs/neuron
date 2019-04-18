package sorts

import (
	"fmt"
	"github.com/neuronlabs/neuron/internal"
	"github.com/neuronlabs/neuron/internal/query/sorts"
	"github.com/neuronlabs/neuron/mapping"
	"net/url"
)

// SortField is a field that contains sorting information
type SortField sorts.SortField

// StructField returns sortfield's structure
func (s *SortField) StructField() *mapping.StructField {
	sField := (*sorts.SortField)(s).StructField()

	return (*mapping.StructField)(sField)
}

// Order returns sortfield's order
func (s *SortField) Order() Order {
	return Order((*sorts.SortField)(s).Order())
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

	if vals, ok := query[internal.QueryParamSort]; ok {

		if len(vals) > 0 {
			v = vals[0]
		}
		if len(v) > 0 {
			v += ","
		}

	}
	v += fmt.Sprintf("%s%s", sign, s.StructField().ApiName())

	query.Set(internal.QueryParamSort, v)

	return query
}
