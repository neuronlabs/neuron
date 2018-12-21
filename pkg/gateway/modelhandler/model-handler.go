package modelhandler

import (
	"errors"
	"fmt"
	"github.com/kucjac/jsonapi/pkg/flags"
	"github.com/kucjac/jsonapi/pkg/gateway/endpoint"
	"github.com/kucjac/jsonapi/pkg/gateway/middlewares"
	"github.com/kucjac/jsonapi/pkg/query"
	"reflect"
)

var IErrInvalidModelEndpoint = errors.New("Invalid model endpoint")

// ModelHandler is the struct that contains all information about given model's endpoints
// It also contains the repository
type ModelHandler struct {
	ModelType reflect.Type

	// Endpoints contains preset information about the provided model.
	Create *endpoint.Endpoint

	Get             *endpoint.Endpoint
	GetRelated      *endpoint.Endpoint
	GetRelationship *endpoint.Endpoint
	List            *endpoint.Endpoint

	Patch             *endpoint.Endpoint
	PatchRelated      *endpoint.Endpoint
	PatchRelationship *endpoint.Endpoint

	Delete *endpoint.Endpoint

	// Flags is the container for the flags variables
	fContainer *flags.Container

	// Repository defines the repository for the provided model
	Repository interface{}
}

// NewModelHandler creates new model handler for given model, with provided repository and with
// support for provided endpoints.
// Returns an error if provided model is not a struct or a ptr to struct or if the endpoint is  of
// unknown type.
func NewModelHandler(
	model interface{},
	repository interface{},
	endpoints ...endpoint.EndpointType,
) (m *ModelHandler, err error) {
	m = new(ModelHandler)
	m.fContainer = flags.New()

	if err = m.newModelType(model); err != nil {
		return
	}
	m.Repository = repository
	for _, e := range endpoints {
		switch e {
		case endpoint.Create:
			m.Create = &endpoint.Endpoint{Type: e}
		case endpoint.Get:
			m.Get = &endpoint.Endpoint{Type: e}
		case endpoint.GetRelated:
			m.GetRelated = &endpoint.Endpoint{Type: e}
		case endpoint.GetRelationship:
			m.GetRelationship = &endpoint.Endpoint{Type: e}
		case endpoint.List:
			m.List = &endpoint.Endpoint{Type: e}
		case endpoint.Patch:
			m.Patch = &endpoint.Endpoint{Type: e}
		case endpoint.Delete:
			m.Delete = &endpoint.Endpoint{Type: e}
		default:
			err = fmt.Errorf("Provided invalid endpoint type for model: %s", m.ModelType.Name())
			return
		}
	}
	return
}

func (m *ModelHandler) Flags() *flags.Container {
	if m.fContainer == nil {
		m.fContainer = flags.New()
	}
	return m.fContainer
}

// AddPresetScope adds preset scope to provided endpoint.
// If the endpoint was not set or is unknown the function returns error.
func (m *ModelHandler) AddPresetPair(
	presetPair *query.PresetPair,
	endpoint endpoint.EndpointType,
) error {
	return m.addPresetPair(presetPair, endpoint, false)
}

// AddPrecheckPair adds the precheck pair to the given model on provided endpoint.
// If the endpoint was not set or is unknown the function returns error.
func (m *ModelHandler) AddPrecheckPair(
	precheckPair *query.PresetPair,
	endpoint endpoint.EndpointType,
) error {
	return m.addPresetPair(precheckPair, endpoint, true)
}

func (m *ModelHandler) AddPresetFilter(
	fieldName string,
	endpointTypes []endpoint.EndpointType,
	operator query.Operator,
	values ...interface{},
) error {
	return nil
}

func (m *ModelHandler) AddPresetSort(
	fieldName string,
	endpointTypes []endpoint.EndpointType,
	order query.Order,
) error {
	return nil
}

func (m *ModelHandler) AddOffsetPresetPaginate(
	limit, offset int,
	endpointTypes []endpoint.EndpointType,
) error {
	return nil
}

// AddEndpoint adds the endpoint to the provided model handler.
// If the endpoint is of unknown type or the handler already contains given endpoint
// an error would be returned.
func (m *ModelHandler) AddEndpoint(endpoint *endpoint.Endpoint) error {
	return m.changeEndpoint(endpoint, false)
}

// AddMiddlewareFunctions adds the middleware functions for given endpoint
func (m *ModelHandler) AddMiddlewareFunctions(
	e endpoint.EndpointType,
	mwares ...middlewares.MiddlewareFunc,
) error {

	var modelEndpoint *endpoint.Endpoint

	switch e {
	case endpoint.Create:
		modelEndpoint = m.Create
	case endpoint.Get:
		modelEndpoint = m.Get
	case endpoint.List:
		modelEndpoint = m.List
	case endpoint.Patch:
		modelEndpoint = m.Patch
	case endpoint.Delete:
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
func (m *ModelHandler) ReplaceEndpoint(endpoint *endpoint.Endpoint) error {
	return m.changeEndpoint(endpoint, true)
}

func (m *ModelHandler) SetAllEndpoints() {
	m.Create = &endpoint.Endpoint{Type: endpoint.Create}
	m.Get = &endpoint.Endpoint{Type: endpoint.Get}
	m.GetRelated = &endpoint.Endpoint{Type: endpoint.GetRelated}
	m.GetRelationship = &endpoint.Endpoint{Type: endpoint.GetRelationship}
	m.List = &endpoint.Endpoint{Type: endpoint.List}
	m.Patch = &endpoint.Endpoint{Type: endpoint.Patch}
	// m.PatchRelated = &endpoint.Endpoint{Type: PatchRelated}
	// m.PatchRelationship = &endpoint.Endpoint{Type: PatchRelationship}
	m.Delete = &endpoint.Endpoint{Type: endpoint.Delete}
}

func (m *ModelHandler) addPresetPair(
	presetPair *query.PresetPair,
	e endpoint.EndpointType,
	check bool,
) error {
	nilEndpoint := func(eName string) error {
		return fmt.Errorf("Adding preset scope on the nil '%s' Endpoint on model: '%s'", eName, m.ModelType.Name())
	}

	switch e {
	case endpoint.Create:
		if m.Create == nil {
			return nilEndpoint("Create")
		}
		if check {
			m.Create.PrecheckPairs = append(m.Create.PrecheckPairs, presetPair)
		} else {
			m.Create.PresetPairs = append(m.Create.PresetPairs, presetPair)
		}

	case endpoint.Get:
		if m.Get == nil {
			return nilEndpoint("Get")
		}
		if check {
			m.Get.PrecheckPairs = append(m.Get.PrecheckPairs, presetPair)
		} else {
			m.Get.PresetPairs = append(m.Get.PresetPairs, presetPair)
		}
	case endpoint.List:
		if m.List == nil {
			return nilEndpoint("List")
		}

		if check {
			m.List.PrecheckPairs = append(m.List.PrecheckPairs, presetPair)
		} else {
			m.List.PresetPairs = append(m.List.PresetPairs, presetPair)
		}

	case endpoint.Patch:
		if m.Patch == nil {
			return nilEndpoint("Patch")
		}

		if check {
			m.Patch.PrecheckPairs = append(m.Patch.PrecheckPairs, presetPair)
		} else {
			m.Patch.PresetPairs = append(m.Patch.PresetPairs, presetPair)
		}

	case endpoint.Delete:
		if m.Delete == nil {
			return nilEndpoint("Delete")
		}

		if check {
			m.Delete.PrecheckPairs = append(m.Delete.PrecheckPairs, presetPair)
		} else {
			m.Delete.PresetPairs = append(m.Delete.PresetPairs, presetPair)
		}

	default:
		return errors.New("Endpoint not specified.")
	}
	return nil
}

func (m *ModelHandler) changeEndpoint(e *endpoint.Endpoint, replace bool) error {
	if e == nil {
		return fmt.Errorf("Provided nil endpoint for model: %s.", m.ModelType.Name())
	}

	switch e.Type {
	case endpoint.Create:
		m.Create = e
	case endpoint.Get:
		m.Get = e
	case endpoint.GetRelated:
		m.GetRelated = e
	case endpoint.GetRelationship:
		m.GetRelationship = e
	case endpoint.List:
		m.List = e
	case endpoint.Patch:
		m.Patch = e
	case endpoint.PatchRelated:
		m.PatchRelated = e
	case endpoint.PatchRelationship:
		m.PatchRelationship = e
	case endpoint.Delete:
		m.Delete = e
	default:
		return IErrInvalidModelEndpoint
	}

	return nil
}

func (m *ModelHandler) newModelType(model interface{}) (err error) {
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		err = errors.New("Invalid model provided. Model must be struct or a pointer to struct.")
		return
	}
	m.ModelType = t
	return
}

/**

Handler Methods with ModelHandler

*/
