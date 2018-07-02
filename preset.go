package jsonapi

import (
	"fmt"
)

// PresetPair is a struct used by presetting / prechecking given model.
type PresetPair struct {
	Scope  *Scope
	Filter *FilterField
	Key    interface{}

	rawPreset string
	rawFilter string
	Error     error
}

func (p *PresetPair) String() string {
	return fmt.Sprintf("Preset: %s, Filter: %s", p.rawPreset, p.rawFilter)
}

func (p *PresetPair) GetPair() (*Scope, *FilterField) {
	scope := p.Scope.copy(true, nil)
	filter := p.Filter.copy()
	return scope, filter
}

// WithKey sets the key value for the given preset pair.
// Used to get value from contexts that matches given pair.
func (p *PresetPair) WithKey(key interface{}) *PresetPair {
	p.Key = key
	return p
}

func (p *PresetPair) ErrOnFail(err error) *PresetPair {
	p.Error = err
	return p
}

func (p *PresetPair) GetStringKey() string {
	if p.Key == nil {
		return ""
	}
	strKey, ok := p.Key.(string)
	if !ok {
		return ""
	}
	return strKey
}

// BuildPresetScope builds the preset scope which should enable the table 'JOIN' feature.
// The query parameter should be URL parseble query ("x1=1&x2=2" etc.)
// The query parameter that are allowed are:
//	- preset=collection.relationfield.relationfield .... - this creates a relation path
//		the last relation field should be of type of model provided as an argument. *REQUIRED
//	- fields[collection]=field1,field2 - describe the fieldset for queries. - used to set
//		the field to get of last preset field. By default the preset field is included into the
//		fieldset. *REQUIRED
//	- filter[collection][field][operator]=value
//	- page[limit][collection] - limit the value of ids within given collection
//	- sort[collection]=field - sorts the collection by provided field. Does not allow nesteds.
// 		@query - url like query that should define how the preset scope should look like. The query
//				allows to set the relation path, filter collections, limit given collection, sort
//				given collection.
//		@fieldFilter - jsonapi field name for provided model the field type must be of the same as
//				the last element of the preset. By default the filter operator is of 'in' type.
//				It must be of form: filter[collection][field]([operator]|([subfield][operator])).
//				The operator or subfield are not required.
func (c *Controller) BuildPresetPair(
	query, fieldFilter string,
) *PresetPair {
	presetScope, filter := c.buildPreparedPair(query, fieldFilter, false)
	return &PresetPair{Scope: presetScope, Filter: filter, rawPreset: query, rawFilter: fieldFilter}
}

// BuildPresetScope builds the preset scope which should enable the table 'JOIN' feature.
// The query parameter should be URL parseble query ("x1=1&x2=2" etc.)
// The query parameter that are allowed are:
//	- preset=collection.relationfield.relationfield .... - this creates a relation path
//		the last relation field should be of type of model provided as an argument. *REQUIRED
//	- fields[collection]=field1,field2 - describe the fieldset for queries. - used to set
//		the field to get of last preset field. By default the preset field is included into the
//		fieldset. *REQUIRED
//	- filter[collection][field][operator]=value
//	- page[limit][collection] - limit the value of ids within given collection
//	- sort[collection]=field - sorts the collection by provided field. Does not allow nesteds.
// 		@query - url like query that should define how the preset scope should look like. The query
//				allows to set the relation path, filter collections, limit given collection, sort
//				given collection.
//		@fieldFilter - jsonapi field name for provided model the field type must be of the same as
//				the last element of the preset. By default the filter operator is of 'in' type.
//				It must be of form: filter[collection][field]([operator]|([subfield][operator])).
//				The operator or subfield are not required.
func (c *Controller) BuildPrecheckPair(
	query, fieldFilter string,
) *PresetPair {
	precheckScope, filter := c.buildPreparedPair(query, fieldFilter, true)
	return &PresetPair{Scope: precheckScope, Filter: filter, rawPreset: query, rawFilter: fieldFilter}
}

// PresetFilter is used to point and filter the target scope with preset values.
// It is a FilterField with a Key field.
// The Key field is used to retrieve the value for given filter from the context.
type PresetFilter struct {
	*FilterField
	Key interface{}
}

// WithKey sets the key value for the given preset filter.
// Used to get value from the context.
func (p *PresetFilter) WithKey(key interface{}) *PresetFilter {
	p.Key = key
	return p
}

// GetStringKey gets the key for given preset filter in the string form.
func (p *PresetFilter) GetStringKey() string {
	if p.Key == nil {
		return ""
	}
	strKey, ok := p.Key.(string)
	if !ok {
		return ""
	}
	return strKey
}

// MustBuildPresetFilter creates new filter field based on the fieldFilter argument and provided
// values. Panics if an error occurs during creation process.
//	Arguments:
//		@fieldFilter - jsonapi field name for provided model the field type must be of the same as
//				the last element of the preset. By default the filter operator is of 'in' type.
//				It must be of form: filter[collection][field]([operator]|([subfield][operator])).
//				The operator or subfield are not required.
//		@values - the values of type matching the filter field
func (c *Controller) MustBuildPresetFilter(
	fieldFilter string,
	values ...interface{},
) *PresetFilter {
	filter, err := c.NewFilterField(fieldFilter, values...)
	if err != nil {
		panic(err)
	}
	return &PresetFilter{FilterField: filter}
}
