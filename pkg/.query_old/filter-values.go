package query

// FilterValues are the values used within the provided FilterField
// It contains the Values which is a slice of provided values for given 'Operator'
type FilterValues struct {
	Values   []interface{}
	Operator *Operator
}

func (f *FilterValues) copy() *FilterValues {
	fv := &FilterValues{Operator: f.Operator}
	fv.Values = make([]interface{}, len(f.Values))
	copy(fv.Values, f.Values)
	return fv
}
