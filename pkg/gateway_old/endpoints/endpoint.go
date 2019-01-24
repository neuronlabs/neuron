package endpoints

import (
	"github.com/kucjac/jsonapi/pkg/flags"
	"github.com/kucjac/jsonapi/pkg/gateway/middlewares"
	"github.com/kucjac/jsonapi/pkg/query/pagination"
	"github.com/kucjac/jsonapi/pkg/query/preset"
	"github.com/kucjac/jsonapi/pkg/query/sorts"
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
	Type EndpointType

	// PrecheckPairs are the pairs of Scope and Filter
	// The scope deines the model from where the preset values should be taken
	// The second defines the filter field for the target model's scope that would be filled with
	// the values from the precheckpair scope
	PrecheckScopes []*preset.Scope

	// PresetPairs are the paris of Scope and jsonapiFilter
	// The scope defines the model from where the preset values should be taken. It should not be
	// the same as target model.
	// The second parameter defines the target model's field and it subfield where the value
	// should be preset.
	PresetScopes []*preset.Scope

	// PresetFilters are the filters for the target model's scope
	// They should be filled with values taken from context with key "PresetFilterValue"
	// When the values are taken the target model's scope would save the value for the relation
	// provided in the filterfield.
	PresetFilters []*preset.Filter

	// PrecheckFilters are the filters for the target model's scope
	// They should be filled with values taken from context with key "PrecheckPairFilterValue"
	// When the values are taken and saved into the precheck filter, the filter is added into the
	// target model's scope.
	PrecheckFilters []*preset.Filter

	// Preset default sorting
	PresetSort []*sorts.SortField

	// Preset default limit offset
	PresetPaginate *pagination.Pagination

	// RelationPrecheckPairs are the prechecks for the GetRelated and GetRelationship root
	RelationPrecheckPairs map[string]*RelationPresetRules

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
	return len(e.PrecheckScopes) > 0 || len(e.PrecheckFilters) > 0
}

// HasPresets checks if the given endpoint has the PresetScopes or PresetFilters
func (e *Endpoint) HasPresets() bool {
	return len(e.PresetScopes) > 0 || len(e.PresetFilters) > 0
}

type RelationPresetRules struct {
	// PrecheckPairs are the pairs of Scope and Filter
	// The scope deines the model from where the preset values should be taken
	// The second defines the filter field for the target model's scope that would be filled with
	// the values from the precheckpair scope
	PrecheckScopes []*preset.Scope

	// PresetPairs are the paris of Scope and jsonapiFilter
	// The scope defines the model from where the preset values should be taken. It should not be
	// the same as target model.
	// The second parameter defines the target model's field and it subfield where the value
	// should be preset.
	PresetScopes []*preset.Scope

	// PresetFilters are the filters for the target model's scope
	// They should be filled with values taken from context with key "PresetFilterValue"
	// When the values are taken the target model's scope would save the value for the relation
	// provided in the filterfield.
	PresetFilters []*preset.Filter

	// PrecheckFilters are the filters for the target model's scope
	// They should be filled with values taken from context with key "PrecheckPairFilterValue"
	// When the values are taken and saved into the precheck filter, the filter is added into the
	// target model's scope.
	PrecheckFilters []*preset.Filter

	// Preset default sorting
	PresetSort []*sorts.SortField

	// Preset default limit offset
	PresetPaginate *pagination.Pagination
}

func (e *Endpoint) String() string {
	return e.Type.String()
}
