package jsonapi

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
)

var IErrInvalidModelEndpoint = errors.New("Invalid model endpoint")

type MiddlewareFunc func(next http.Handler) http.Handler

// ModelHandler defines how the
type ModelHandler struct {
	ModelType reflect.Type

	// Endpoints contains preset information about the provided model.
	Create *Endpoint

	Get             *Endpoint
	GetRelated      *Endpoint
	GetRelationship *Endpoint
	List            *Endpoint

	Patch             *Endpoint
	PatchRelated      *Endpoint
	PatchRelationship *Endpoint

	Delete *Endpoint

	// FlagUseLinks is a flag that defines if the endpoint should use links
	FlagUseLinks *bool

	// GetModified defines if the result for Patch Should be returned.
	FlagReturnPatchContent *bool

	// FlagMetaCountList is a flag that defines if the List result should include objects count
	FlagMetaCountList *bool

	// Repository defines the repository for the provided model
	Repository Repository
}

type ModelPresetGetter interface {
	GetPresetPair(endpoint EndpointType, controller *Controller) *PresetPair
}

type ModelPrecheckGetter interface {
	GetPrecheckPair(endpoint EndpointType, controller *Controller) *PresetPair
}

// NewModelHandler creates new model handler for given model, with provided repository and with
// support for provided endpoints.
// Returns an error if provided model is not a struct or a ptr to struct or if the endpoint is  of
// unknown type.
func NewModelHandler(
	model interface{},
	repository Repository,
	endpoints ...EndpointType,
) (m *ModelHandler, err error) {
	m = new(ModelHandler)

	if err = m.newModelType(model); err != nil {
		return
	}
	m.Repository = repository
	for _, endpoint := range endpoints {
		switch endpoint {
		case Create:
			m.Create = &Endpoint{Type: endpoint}
		case Get:
			m.Get = &Endpoint{Type: endpoint}
		case GetRelated:
			m.GetRelated = &Endpoint{Type: endpoint}
		case GetRelationship:
			m.GetRelationship = &Endpoint{Type: endpoint}
		case List:
			m.List = &Endpoint{Type: endpoint}
		case Patch:
			m.Patch = &Endpoint{Type: endpoint}
		case Delete:
			m.Delete = &Endpoint{Type: endpoint}
		default:
			err = fmt.Errorf("Provided invalid endpoint type for model: %s", m.ModelType.Name())
			return
		}
	}
	return
}

// AddPresetScope adds preset scope to provided endpoint.
// If the endpoint was not set or is unknown the function returns error.
func (m *ModelHandler) AddPresetPair(
	presetPair *PresetPair,
	endpoint EndpointType,
) error {
	return m.addPresetPair(presetPair, endpoint, false)
}

// AddPrecheckPair adds the precheck pair to the given model on provided endpoint.
// If the endpoint was not set or is unknown the function returns error.
func (m *ModelHandler) AddPrecheckPair(
	precheckPair *PresetPair,
	endpoint EndpointType,
) error {
	return m.addPresetPair(precheckPair, endpoint, true)
}

func (m *ModelHandler) AddPresetFilter(
	fieldName string,
	endpointTypes []EndpointType,
	operator FilterOperator,
	values ...interface{},
) error {
	return nil
}

func (m *ModelHandler) AddPresetSort(
	fieldName string,
	endpointTypes []EndpointType,
	order Order,
) error {
	return nil
}

func (m *ModelHandler) AddOffsetPresetPaginate(
	limit, offset int,
	endpointTypes []EndpointType,
) error {
	return nil
}

// AddEndpoint adds the endpoint to the provided model handler.
// If the endpoint is of unknown type or the handler already contains given endpoint
// an error would be returned.
func (m *ModelHandler) AddEndpoint(endpoint *Endpoint) error {
	return m.changeEndpoint(endpoint, false)
}

// AddMiddlewareFunctions adds the middleware functions for given endpoint
func (m *ModelHandler) AddMiddlewareFunctions(endpoint EndpointType, middlewares ...MiddlewareFunc) error {
	var modelEndpoint *Endpoint
	switch endpoint {
	case Create:
		modelEndpoint = m.Create
	case Get:
		modelEndpoint = m.Get
	case List:
		modelEndpoint = m.List
	case Patch:
		modelEndpoint = m.Patch
	case Delete:
		modelEndpoint = m.Delete
	}

	if modelEndpoint == nil {
		err := fmt.Errorf("Invalid endpoint provided: %v", endpoint)
		return err
	}

	modelEndpoint.Middlewares = append(modelEndpoint.Middlewares, middlewares...)
	return nil
}

// ReplaceEndpoint replaces the endpoint for the provided model handler.
// If the endpoint is of unknown type the function returns an error.
func (m *ModelHandler) ReplaceEndpoint(endpoint *Endpoint) error {
	return m.changeEndpoint(endpoint, true)
}

func (m *ModelHandler) SetAllEndpoints() {
	m.Create = &Endpoint{Type: Create}
	m.Get = &Endpoint{Type: Get}
	m.GetRelated = &Endpoint{Type: GetRelated}
	m.GetRelationship = &Endpoint{Type: GetRelationship}
	m.List = &Endpoint{Type: List}
	m.Patch = &Endpoint{Type: Patch}
	// m.PatchRelated = &Endpoint{Type: PatchRelated}
	// m.PatchRelationship = &Endpoint{Type: PatchRelationship}
	m.Delete = &Endpoint{Type: Delete}
}

func (m *ModelHandler) addPresetPair(
	presetPair *PresetPair,
	endpoint EndpointType,
	check bool,
) error {
	nilEndpoint := func(eName string) error {
		return fmt.Errorf("Adding preset scope on the nil '%s' Endpoint on model: '%s'", eName, m.ModelType.Name())
	}

	switch endpoint {
	case Create:
		if m.Create == nil {
			return nilEndpoint("Create")
		}
		if check {
			m.Create.PrecheckPairs = append(m.Create.PrecheckPairs, presetPair)
		} else {
			m.Create.PresetPairs = append(m.Create.PresetPairs, presetPair)
		}

	case Get:
		if m.Get == nil {
			return nilEndpoint("Get")
		}
		if check {
			m.Get.PrecheckPairs = append(m.Get.PrecheckPairs, presetPair)
		} else {
			m.Get.PresetPairs = append(m.Get.PresetPairs, presetPair)
		}
	case List:
		if m.List == nil {
			return nilEndpoint("List")
		}

		if check {
			m.List.PrecheckPairs = append(m.List.PrecheckPairs, presetPair)
		} else {
			m.List.PresetPairs = append(m.List.PresetPairs, presetPair)
		}

	case Patch:
		if m.Patch == nil {
			return nilEndpoint("Patch")
		}

		if check {
			m.Patch.PrecheckPairs = append(m.Patch.PrecheckPairs, presetPair)
		} else {
			m.Patch.PresetPairs = append(m.Patch.PresetPairs, presetPair)
		}

	case Delete:
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

func (m *ModelHandler) changeEndpoint(endpoint *Endpoint, replace bool) error {
	if endpoint == nil {
		return fmt.Errorf("Provided nil endpoint for model: %s.", m.ModelType.Name())
	}

	switch endpoint.Type {
	case Create:
		m.Create = endpoint
	case Get:
		m.Get = endpoint
	case GetRelated:
		m.GetRelated = endpoint
	case GetRelationship:
		m.GetRelationship = endpoint
	case List:
		m.List = endpoint
	case Patch:
		m.Patch = endpoint
	case PatchRelated:
		m.PatchRelated = endpoint
	case PatchRelationship:
		m.PatchRelationship = endpoint
	case Delete:
		m.Delete = endpoint
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

// AddModelsPresetPair gets the model handler from the Handler, and adds the presetpair
// to the specific endpoint for this model.
// Returns error if the model is not present within Handler or if the model does not
// support given endpoint.
func (h *Handler) AddModelsPresetPair(
	model interface{},
	presetPair *PresetPair,
	endpoint EndpointType,
) error {
	handler, err := h.getModelHandler(model)
	if err != nil {
		return err
	}

	if err := handler.AddPresetPair(presetPair, endpoint); err != nil {
		return err
	}
	return nil
}

// AddModelsPrecheckPair gets the model handler from the Handler, and adds the precheckPair
// to the specific endpoint for this model.
// Returns error if the model is not present within Handler or if the model does not
// support given endpoint.
func (h *Handler) AddModelsPrecheckPair(
	model interface{},
	precheckPair *PresetPair,
	endpoint EndpointType,
) error {
	handler, err := h.getModelHandler(model)
	if err != nil {
		return err
	}

	if err := handler.AddPresetPair(precheckPair, endpoint); err != nil {
		return err
	}
	return nil
}

// AddModelsEndpoint adds the endpoint to the provided model.
// If the model is not set within given handler, an endpoint is already occupied or is of unknown
// type the function returns error.
func (h *Handler) AddModelsEndpoint(model interface{}, endpoint *Endpoint) error {
	handler, err := h.getModelHandler(model)
	if err != nil {
		return err
	}
	if err := handler.AddEndpoint(endpoint); err != nil {
		return err
	}
	return nil
}

// ReplaceModelsEndpoint replaces an endpoint for provided model.
// If the model is not set within Handler or an endpoint is of unknown type the function
// returns an error.
func (h *Handler) ReplaceModelsEndpoint(model interface{}, endpoint *Endpoint) error {
	handler, err := h.getModelHandler(model)
	if err != nil {
		return err
	}
	if err := handler.ReplaceEndpoint(endpoint); err != nil {
		return err
	}
	return nil
}

// GetModelHandler gets the model handler that matches the provided model type.
// If no handler is found within Handler the function returns an error.
func (h *Handler) GetModelHandler(model interface{}) (mHandler *ModelHandler, err error) {
	return h.getModelHandler(model)
}

func (h *Handler) getModelHandler(model interface{}) (mHandler *ModelHandler, err error) {
	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}
	var ok bool
	mHandler, ok = h.ModelHandlers[modelType]
	if !ok {
		err = IErrModelHandlerNotFound
		return
	}
	return

}
