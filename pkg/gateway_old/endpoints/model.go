package endpoints

import (
	"errors"
	"fmt"
	"github.com/kucjac/jsonapi/pkg/flags"
	"github.com/kucjac/jsonapi/pkg/gateway/middlewares"
	"github.com/kucjac/jsonapi/pkg/mapping"
	"github.com/kucjac/jsonapi/pkg/query/filters"
	"github.com/kucjac/jsonapi/pkg/query/preset"
	"github.com/kucjac/jsonapi/pkg/query/sorts"
)

var IErrInvalidModelEndpoint = errors.New("Invalid model endpoint")

// Model is the struct that contains all information about given model's endpoints
// It also contains the repository
type Model struct {
	// mStruct defines the Structure for given model
	mStruct *mapping.ModelStruct

	// Endpoints contains preset information about the provided model.
	Create *Endpoint

	Get             *Endpoint
	GetRelated      *Endpoint
	GetRelationship *Endpoint
	List            *Endpoint

	Patch             *Endpoint
	PatchRelationship *Endpoint

	Delete *Endpoint

	// Flags is the container for the flags variables
	fContainer *flags.Container

	// Repository defines the repository for the provided model
	Repository interface{}
}

func (m *Model) Flags() *flags.Container {
	if m.fContainer == nil {
		m.fContainer = flags.New()
	}
	return m.fContainer
}

// AddPresetScope adds preset scope to provided Tp
// If the endpoint was not set or is unknown the function returns error.
func (m *Model) AddPresetPair(
	presetPair *preset.Scope,
	endpoint EndpointType,
) error {
	return m.addPresetPair(presetPair, endpoint, false)
}

// AddPrecheckPair adds the precheck pair to the given model on provided Tp
// If the endpoint was not set or is unknown the function returns error.
func (m *Model) AddPrecheckPair(
	precheckPair *preset.Scope,
	endpoint EndpointType,
) error {
	return m.addPresetPair(precheckPair, endpoint, true)
}

func (m *Model) AddPresetFilter(
	fieldName string,
	endpointTypes []EndpointType,
	operator *filters.Operator,
	values ...interface{},
) error {
	return nil
}

func (m *Model) AddPresetSort(
	fieldName string,
	endpointTypes []EndpointType,
	order sorts.Order,
) error {
	return nil
}

func (m *Model) AddOffsetPresetPaginate(
	limit, offset int,
	endpointTypes []EndpointType,
) error {
	return nil
}

// AddEndpoint adds the endpoint to the provided model handler.
// If the endpoint is of unknown type or the handler already contains given endpoint
// an error would be returned.
func (m *Model) AddEndpoint(endpoint *Endpoint) error {
	return m.changeEndpoint(endpoint, false)
}

// AddMiddlewareFunctions adds the middleware functions for given endpoint
func (m *Model) AddMiddlewareFunctions(
	e EndpointType,
	mwares ...middlewares.MiddlewareFunc,
) error {

	var modelEndpoint *Endpoint

	switch e {
	case TpCreate:
		modelEndpoint = m.Create
	case TpGet:
		modelEndpoint = m.Get
	case TpList:
		modelEndpoint = m.List
	case TpPatch:
		modelEndpoint = m.Patch
	case TpDelete:
		modelEndpoint = m.Delete
	}

	if modelEndpoint == nil {
		err := fmt.Errorf("Invalid endpoint provided: %v", e)
		return err
	}

	modelEndpoint.Middlewares = append(modelEndpoint.Middlewares, mwares...)
	return nil
}

// ReplaceEndpoint replaces the endpoint for the provided model handler.
// If the endpoint is of unknown type the function returns an error.
func (m *Model) ReplaceEndpoint(endpoint *Endpoint) error {
	return m.changeEndpoint(endpoint, true)
}

func (m *Model) SetAllEndpoints() {
	m.Create = &Endpoint{Type: TpCreate}
	m.Get = &Endpoint{Type: TpGet}
	m.GetRelated = &Endpoint{Type: TpGetRelated}
	m.GetRelationship = &Endpoint{Type: TpGetRelationship}
	m.List = &Endpoint{Type: TpList}
	m.Patch = &Endpoint{Type: TpPatch}
	// m.PatchRelated = &Endpoint{Type: PatchRelated}
	// m.PatchRelationship = &Endpoint{Type: PatchRelationship}
	m.Delete = &Endpoint{Type: TpDelete}
}

func (m *Model) addPresetPair(
	presetPair *preset.Scope,
	e EndpointType,
	check bool,
) error {
	nilEndpoint := func(eName string) error {
		return fmt.Errorf("Adding preset scope on the nil '%s' Endpoint on model: '%s'", eName, m.mStruct.Type().Name())
	}

	switch e {
	case TpCreate:
		if m.Create == nil {
			return nilEndpoint("Create")
		}
		if check {
			m.Create.PrecheckScopes = append(m.Create.PrecheckScopes, presetPair)
		} else {
			m.Create.PresetScopes = append(m.Create.PresetScopes, presetPair)
		}

	case TpGet:
		if m.Get == nil {
			return nilEndpoint("Get")
		}
		if check {
			m.Get.PrecheckScopes = append(m.Get.PrecheckScopes, presetPair)
		} else {
			m.Get.PresetScopes = append(m.Get.PresetScopes, presetPair)
		}
	case TpList:
		if m.List == nil {
			return nilEndpoint("List")
		}

		if check {
			m.List.PrecheckScopes = append(m.List.PrecheckScopes, presetPair)
		} else {
			m.List.PresetScopes = append(m.List.PresetScopes, presetPair)
		}

	case TpPatch:
		if m.Patch == nil {
			return nilEndpoint("Patch")
		}

		if check {
			m.Patch.PrecheckScopes = append(m.Patch.PrecheckScopes, presetPair)
		} else {
			m.Patch.PresetScopes = append(m.Patch.PresetScopes, presetPair)
		}

	case TpDelete:
		if m.Delete == nil {
			return nilEndpoint("Delete")
		}

		if check {
			m.Delete.PrecheckScopes = append(m.Delete.PrecheckScopes, presetPair)
		} else {
			m.Delete.PresetScopes = append(m.Delete.PresetScopes, presetPair)
		}

	default:
		return errors.New("Endpoint not specified.")
	}
	return nil
}

func (m *Model) changeEndpoint(e *Endpoint, replace bool) error {
	if e == nil {
		return fmt.Errorf("Provided nil endpoint for model: %s.", m.mStruct.Type().Name())
	}

	switch e.Type {
	case TpCreate:
		m.Create = e
	case TpGet:
		m.Get = e
	case TpGetRelated:
		m.GetRelated = e
	case TpGetRelationship:
		m.GetRelationship = e
	case TpList:
		m.List = e
	case TpPatch:
		m.Patch = e
	case TpPatchRelationship:
		m.PatchRelationship = e
	case TpDelete:
		m.Delete = e
	default:
		return IErrInvalidModelEndpoint
	}

	return nil
}
