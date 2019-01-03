package controller

import (
	"github.com/kucjac/jsonapi/pkg/query"
)

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
) *query.PresetPair {
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
) *query.PresetPair {
	precheckScope, filter := c.buildPreparedPair(query, fieldFilter, true)
	return &query.PresetPair{Scope: precheckScope, Filter: filter, rawPreset: query, rawFilter: fieldFilter}
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
) *query.PresetFilter {
	filter, err := c.NewFilterFieldWithForeigns(fieldFilter, values...)
	if err != nil {
		panic(err)
	}
	return &query.PresetFilter{query.FilterField: filter}
}
