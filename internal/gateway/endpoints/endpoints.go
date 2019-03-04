package endpoints

import (
	"github.com/kucjac/jsonapi/flags"
	"github.com/kucjac/jsonapi/gateway/middlewares"
	"github.com/kucjac/jsonapi/internal/query/paginations"
	"github.com/kucjac/jsonapi/internal/query/preset"
	"github.com/kucjac/jsonapi/internal/query/sorts"
	"net/http"
)

type EndpointType int

const (
	TpUnknown EndpointType = iota

	// Create
	TpCreate

	// Gets
	TpGet
	TpGetRelated
	TpGetRelationship

	// List
	TpList

	// Patches
	TpPatch
	TpPatchRelated
	TpPatchRelationship

	// Deletes
	TpDelete
)

func (e EndpointType) String() string {
	var op string
	switch e {
	case TpCreate:
		op = "CREATE"
	case TpGet:
		op = "GET"
	case TpGetRelated:
		op = "GET RELATED"
	case TpGetRelationship:
		op = "GET RELATIONSHIP"
	case TpList:
		op = "LIST"
	case TpPatch:
		op = "PATCH"
	case TpPatchRelationship:
		op = "PATCH RELATIONSHIP"
	case TpDelete:
		op = "DELETE"

	default:
		op = "UNKNOWN"
	}
	return op
}

type Endpoint struct {
	// Type is the endpoint type
	tp EndpointType

	// PrecheckPairs are the pairs of Scope and Filter
	// The scope deines the model from where the preset values should be taken
	// The second defines the filter field for the target model's scope that would be filled with
	// the values from the precheckpair scope
	precheckScopes []*preset.Scope

	// PresetPairs are the paris of Scope and jsonapiFilter
	// The scope defines the model from where the preset values should be taken. It should not be
	// the same as target model.
	// The second parameter defines the target model's field and it subfield where the value
	// should be preset.
	presetScopes []*preset.Scope

	// PresetFilters are the filters for the target model's scope
	// They should be filled with values taken from context with key "PresetFilterValue"
	// When the values are taken the target model's scope would save the value for the relation
	// provided in the filterfield.
	presetFilters []*preset.Filter

	// PrecheckFilters are the filters for the target model's scope
	// They should be filled with values taken from context with key "PrecheckPairFilterValue"
	// When the values are taken and saved into the precheck filter, the filter is added into the
	// target model's scope.
	precheckFilters []*preset.Filter

	// Preset default sorting
	presetSort []*sorts.SortField

	// Preset default limit offset
	presetPaginate *paginations.Pagination

	// RelationPrecheckPairs are the prechecks for the GetRelated and GetRelationship root
	relationPrecheckPairs map[string]*RelationPresetRules

	Middlewares []middlewares.MiddlewareFunc

	// Flags are the default behavior definitions for given endpoint
	container *flags.Container

	// CustomHandlerFunc is a http.HandlerFunc defined for this endpoint
	customHandlerFunc http.HandlerFunc
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
	return len(e.precheckScopes) > 0 || len(e.precheckFilters) > 0
}

// HasPresets checks if the given endpoint has the presetPairs or presetFilters
func (e *Endpoint) HasPresets() bool {
	return len(e.presetScopes) > 0 || len(e.presetFilters) > 0
}

type RelationPresetRules struct {
	// PrecheckPairs are the pairs of Scope and Filter
	// The scope deines the model from where the preset values should be taken
	// The second defines the filter field for the target model's scope that would be filled with
	// the values from the precheckpair scope
	precheckPairs []*preset.Scope

	// PresetPairs are the paris of Scope and jsonapiFilter
	// The scope defines the model from where the preset values should be taken. It should not be
	// the same as target model.
	// The second parameter defines the target model's field and it subfield where the value
	// should be preset.
	presetPairs []*preset.Scope

	// PresetFilters are the filters for the target model's scope
	// They should be filled with values taken from context with key "PresetFilterValue"
	// When the values are taken the target model's scope would save the value for the relation
	// provided in the filterfield.
	presetFilters []*preset.Filter

	// PrecheckFilters are the filters for the target model's scope
	// They should be filled with values taken from context with key "PrecheckPairFilterValue"
	// When the values are taken and saved into the precheck filter, the filter is added into the
	// target model's scope.
	precheckFilters []*preset.Filter

	// Preset default sorting
	presetSort []*sorts.SortField

	// Preset default limit offset
	presetPaginate *paginations.Pagination
}

func (e *Endpoint) String() string {
	return e.tp.String()
}
