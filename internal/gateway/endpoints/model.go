package endpoints

import (
	"fmt"
	"github.com/kucjac/jsonapi/flags"
	"github.com/kucjac/jsonapi/gateway/middlewares"
	"github.com/kucjac/jsonapi/internal/models"
	"github.com/kucjac/jsonapi/internal/query/filters"
	"github.com/kucjac/jsonapi/internal/query/preset"
	"github.com/kucjac/jsonapi/internal/query/sorts"
	"github.com/kucjac/jsonapi/log"
	"github.com/pkg/errors"
)

var ErrInvalidModelEndpoint = errors.New("Invalid model endpoint")

// Model is the struct that contains all information about given model's endpoints
// It also contains the repository
type Model struct {
	// mStruct defines the Structure for given model
	mStruct *models.ModelStruct

	// Endpoints contains preset information about the provided model.
	create *Endpoint

	get             *Endpoint
	getRelated      *Endpoint
	getRelationship *Endpoint
	list            *Endpoint

	patch             *Endpoint
	patchRelationship *Endpoint

	deleteEp *Endpoint

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

// // AddPrecheckPair adds the precheck pair to the given model on provided Tp
// // If the endpoint was not set or is unknown the function returns error.
// func (m *Model) AddPrecheckPair(
// 	precheckPair *preset.Scope,
// 	endpoint EndpointType,
// ) error {
// 	return m.addPresetPair(precheckPair, endpoint, true)
// }

// func (m *Model) AddPresetFilter(
// 	fieldName string,
// 	endpointTypes []EndpointType,
// 	operator *filters.Operator,
// 	values ...interface{},
// ) error {
// 	return nil
// }

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

// SetEndpoint sets the endpoint by it's type
func (m *Model) SetEndpoint(e *Endpoint) error {
	switch e.tp {
	case TpCreate:
		m.create = e
	case TpGet:
		m.get = e
	case TpGetRelated:
		m.getRelated = e
	case TpGetRelationship:
		m.getRelationship = e
	case TpList:
		m.list = e
	case TpPatch:
		m.patch = e
	case TpDelete:
		m.deleteEp = e
	default:
		log.Errorf("%v. Unknown endpoint type. Model: %v, Endpoint: %+v", ErrInvalidModelEndpoint, m.mStruct.Type(), e)
		return ErrInvalidModelEndpoint
	}
	return nil
}

// // AddEndpoint adds the endpoint to the provided model handler.
// // If the endpoint is of unknown type or the handler already contains given endpoint
// // an error would be returned.
// func (m *Model) AddEndpoint(endpoint *Endpoint) error {
// 	return m.changeEndpoint(endpoint, false)
// }

// // AddMiddlewareFunctions adds the middleware functions for given endpoint
// func (m *Model) AddMiddlewareFunctions(
// 	e EndpointType,
// 	mwares ...middlewares.MiddlewareFunc,
// ) error {

// 	var modelEndpoint *Endpoint

// 	switch e {
// 	case TpCreate:
// 		modelEndpoint = m.create
// 	case TpGet:
// 		modelEndpoint = m.get
// 	case TpList:
// 		modelEndpoint = m.list
// 	case TpPatch:
// 		modelEndpoint = m.patch
// 	case TpDelete:
// 		modelEndpoint = m.deleteEp
// 	}

// 	if modelEndpoint == nil {
// 		err := fmt.Errorf("Invalid endpoint provided: %v", e)
// 		return err
// 	}

// 	modelEndpoint.Middlewares = append(modelEndpoint.Middlewares, mwares...)
// 	return nil
// }

// // ReplaceEndpoint replaces the endpoint for the provided model handler.
// // If the endpoint is of unknown type the function returns an error.
// func (m *Model) ReplaceEndpoint(endpoint *Endpoint) error {
// 	return m.changeEndpoint(endpoint, true)
// }

// // func (m *Model) SetAllEndpoints() {
// // 	m.create = &Endpoint{tp: TpCreate}
// // 	m.get = &Endpoint{tp: TpGet}
// // 	m.getRelated = &Endpoint{tp: TpGetRelated}
// // 	m.getRelationship = &Endpoint{tp: TpGetRelationship}
// // 	m.list = &Endpoint{tp: TpList}
// // 	m.patch = &Endpoint{tp: TpPatch}
// // 	// m.PatchRelated = &Endpoint{Type: PatchRelated}
// // 	// m.PatchRelationship = &Endpoint{Type: PatchRelationship}
// // 	m.deleteEp = &Endpoint{tp: TpDelete}
// // }

// func (m *Model) addPresetPair(
// 	presetPair *preset.Scope,
// 	e EndpointType,
// 	check bool,
// ) error {
// 	nilEndpoint := func(eName string) error {
// 		return fmt.Errorf("Adding preset scope on the nil '%s' Endpoint on model: '%s'", eName, m.mStruct.Type().Name())
// 	}

// 	switch e {
// 	case TpCreate:
// 		if m.create == nil {
// 			return nilEndpoint("Create")
// 		}
// 		if check {
// 			m.create.precheckScopes = append(m.create.precheckScopes, presetPair)
// 		} else {
// 			m.create.presetScopes = append(m.create.presetScopes, presetPair)
// 		}

// 	case TpGet:
// 		if m.get == nil {
// 			return nilEndpoint("Get")
// 		}
// 		if check {
// 			m.get.precheckScopes = append(m.get.precheckScopes, presetPair)
// 		} else {
// 			m.get.presetScopes = append(m.get.presetScopes, presetPair)
// 		}
// 	case TpList:
// 		if m.list == nil {
// 			return nilEndpoint("List")
// 		}

// 		if check {
// 			m.list.precheckScopes = append(m.list.precheckScopes, presetPair)
// 		} else {
// 			m.list.presetScopes = append(m.list.presetScopes, presetPair)
// 		}

// 	case TpPatch:
// 		if m.patch == nil {
// 			return nilEndpoint("Patch")
// 		}

// 		if check {
// 			m.patch.precheckScopes = append(m.patch.precheckScopes, presetPair)
// 		} else {
// 			m.patch.presetScopes = append(m.patch.presetScopes, presetPair)
// 		}

// 	case TpDelete:
// 		if m.deleteEp == nil {
// 			return nilEndpoint("Delete")
// 		}

// 		if check {
// 			m.deleteEp.precheckScopes = append(m.deleteEp.precheckScopes, presetPair)
// 		} else {
// 			m.deleteEp.presetScopes = append(m.deleteEp.presetScopes, presetPair)
// 		}

// 	default:
// 		return errors.New("Endpoint not specified.")
// 	}
// 	return nil
// }

func (m *Model) changeEndpoint(e *Endpoint, replace bool) error {
	if e == nil {
		return fmt.Errorf("Provided nil endpoint for model: %s.", m.mStruct.Type().Name())
	}

	switch e.tp {
	case TpCreate:
		m.create = e
	case TpGet:
		m.get = e
	case TpGetRelated:
		m.getRelated = e
	case TpGetRelationship:
		m.getRelationship = e
	case TpList:
		m.list = e
	case TpPatch:
		m.patch = e
	case TpPatchRelationship:
		m.patchRelationship = e
	case TpDelete:
		m.deleteEp = e
	default:
		return IErrInvalidModelEndpoint
	}

	return nil
}
