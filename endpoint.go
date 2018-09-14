package jsonapi

import (
	"net/http"
)

type EndpointType int

const (
	UnkownPath EndpointType = iota

	// Create
	Create

	// Gets
	Get
	GetRelated
	GetRelationship

	// List
	List

	// Patches
	Patch
	PatchRelated
	PatchRelationship

	// Deletes
	Delete
)

func (e EndpointType) String() string {
	var op string
	switch e {
	case Create:
		op = "CREATE"
	case Get:
		op = "GET"
	case GetRelated:
		op = "GET RELATED"
	case GetRelationship:
		op = "GET RELATIONSHIP"
	case List:
		op = "LIST"
	case Patch:
		op = "PATCH"
	case PatchRelationship:
		op = "PATCH RELATIONSHIP"
	case Delete:
		op = "DELETE"

	default:
		op = "UNKNOWN"
	}
	return op
}

type Endpoint struct {
	// Type is the endpoint type
	Type EndpointType

	// PrecheckPairs are the pairs of Scope and Filter
	// The scope deines the model from where the preset values should be taken
	// The second defines the filter field for the target model's scope that would be filled with
	// the values from the precheckpair scope
	PrecheckPairs []*PresetPair

	// PresetPairs are the paris of Scope and jsonapiFilter
	// The scope defines the model from where the preset values should be taken. It should not be
	// the same as target model.
	// The second parameter defines the target model's field and it subfield where the value
	// should be preset.
	PresetPairs []*PresetPair

	// PresetFilters are the filters for the target model's scope
	// They should be filled with values taken from context with key "PresetFilterValue"
	// When the values are taken the target model's scope would save the value for the relation
	// provided in the filterfield.
	PresetFilters []*PresetFilter

	// PrecheckFilters are the filters for the target model's scope
	// They should be filled with values taken from context with key "PrecheckPairFilterValue"
	// When the values are taken and saved into the precheck filter, the filter is added into the
	// target model's scope.
	PrecheckFilters []*PresetFilter

	// Preset default sorting
	PresetSort []*SortField

	// Preset default limit offset
	PresetPaginate *Pagination

	// RelationPrecheckPairs are the prechecks for the GetRelated and GetRelationship root
	RelationPrecheckPairs map[string]*RelationPresetRules

	Middlewares []MiddlewareFunc

	// FlagUseLinks is a flag that defines if the endpoint should use links
	FlagUseLinks *bool

	// GetModified defines if the result for Patch Should be returned.
	FlagReturnPatchContent *bool

	// FlagMetaCountList is a flag that defines if the List result should include objects count
	FlagMetaCountList *bool

	// CustomHandlerFunc is a http.HandlerFunc defined for this endpoint
	CustomHandlerFunc http.HandlerFunc
}

func (e *Endpoint) HasPrechecks() bool {
	return len(e.PrecheckFilters) > 0 || len(e.PrecheckFilters) > 0
}

func (e *Endpoint) HasPresets() bool {
	return len(e.PresetPairs) > 0 || len(e.PresetFilters) > 0
}

type RelationPresetRules struct {
	// PrecheckPairs are the pairs of Scope and Filter
	// The scope deines the model from where the preset values should be taken
	// The second defines the filter field for the target model's scope that would be filled with
	// the values from the precheckpair scope
	PrecheckPairs []*PresetPair

	// PresetPairs are the paris of Scope and jsonapiFilter
	// The scope defines the model from where the preset values should be taken. It should not be
	// the same as target model.
	// The second parameter defines the target model's field and it subfield where the value
	// should be preset.
	PresetPairs []*PresetPair

	// PresetFilters are the filters for the target model's scope
	// They should be filled with values taken from context with key "PresetFilterValue"
	// When the values are taken the target model's scope would save the value for the relation
	// provided in the filterfield.
	PresetFilters []*PresetFilter

	// PrecheckFilters are the filters for the target model's scope
	// They should be filled with values taken from context with key "PrecheckPairFilterValue"
	// When the values are taken and saved into the precheck filter, the filter is added into the
	// target model's scope.
	PrecheckFilters []*PresetFilter

	// Preset default sorting
	PresetSort []*SortField

	// Preset default limit offset
	PresetPaginate *Pagination
}

func (e *Endpoint) String() string {
	return e.Type.String()
}
