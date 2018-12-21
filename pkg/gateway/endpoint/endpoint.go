package endpoint

import (
	"github.com/kucjac/jsonapi/pkg/flags"
	"github.com/kucjac/jsonapi/pkg/gateway/middlewares"
	"github.com/kucjac/jsonapi/pkg/query"
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
	PrecheckPairs []*query.PresetPair

	// PresetPairs are the paris of Scope and jsonapiFilter
	// The scope defines the model from where the preset values should be taken. It should not be
	// the same as target model.
	// The second parameter defines the target model's field and it subfield where the value
	// should be preset.
	PresetPairs []*query.PresetPair

	// PresetFilters are the filters for the target model's scope
	// They should be filled with values taken from context with key "PresetFilterValue"
	// When the values are taken the target model's scope would save the value for the relation
	// provided in the filterfield.
	PresetFilters []*query.PresetFilter

	// PrecheckFilters are the filters for the target model's scope
	// They should be filled with values taken from context with key "PrecheckPairFilterValue"
	// When the values are taken and saved into the precheck filter, the filter is added into the
	// target model's scope.
	PrecheckFilters []*query.PresetFilter

	// Preset default sorting
	PresetSort []*query.SortField

	// Preset default limit offset
	PresetPaginate *query.Pagination

	// RelationPrecheckPairs are the prechecks for the GetRelated and GetRelationship root
	RelationPrecheckPairs map[string]*RelationPresetRules

	Middlewares []middlewares.MiddlewareFunc

	// Flags are the default behavior definitions for given endpoint
	container *flags.Container

	// CustomHandlerFunc is a http.HandlerFunc defined for this endpoint
	CustomHandlerFunc http.HandlerFunc
}

// Flags returns the flags related to the given endpoint
func (e *Endpoint) Flags() *flags.Container {
	if e.container == nil {
		e.container = flags.New()
	}
	return e.container
}

// HasPrechecks checks if the precheck values are set on the given endpoint
func (e *Endpoint) HasPrechecks() bool {
	return len(e.PrecheckFilters) > 0 || len(e.PrecheckFilters) > 0
}

// HasPresets checks if the given endpoint has the presetPairs or presetFilters
func (e *Endpoint) HasPresets() bool {
	return len(e.PresetPairs) > 0 || len(e.PresetFilters) > 0
}

type RelationPresetRules struct {
	// PrecheckPairs are the pairs of Scope and Filter
	// The scope deines the model from where the preset values should be taken
	// The second defines the filter field for the target model's scope that would be filled with
	// the values from the precheckpair scope
	PrecheckPairs []*query.PresetPair

	// PresetPairs are the paris of Scope and jsonapiFilter
	// The scope defines the model from where the preset values should be taken. It should not be
	// the same as target model.
	// The second parameter defines the target model's field and it subfield where the value
	// should be preset.
	PresetPairs []*query.PresetPair

	// PresetFilters are the filters for the target model's scope
	// They should be filled with values taken from context with key "PresetFilterValue"
	// When the values are taken the target model's scope would save the value for the relation
	// provided in the filterfield.
	PresetFilters []*query.PresetFilter

	// PrecheckFilters are the filters for the target model's scope
	// They should be filled with values taken from context with key "PrecheckPairFilterValue"
	// When the values are taken and saved into the precheck filter, the filter is added into the
	// target model's scope.
	PrecheckFilters []*query.PresetFilter

	// Preset default sorting
	PresetSort []*query.SortField

	// Preset default limit offset
	PresetPaginate *query.Pagination
}

func (e *Endpoint) String() string {
	return e.Type.String()
}
