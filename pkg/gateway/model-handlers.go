package gateway

// AddModelsPresetPair gets the model handler from the Handler, and adds the presetpair
// to the specific endpoint for this model.
// Returns error if the model is not present within Handler or if the model does not
// support given endpoint.
func (h *Handler) AddModelsPresetPair(
	model interface{},
	presetPair *query.PresetPair,
	endpoint endpoint.EndpointType,
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
	precheckPair *query.PresetPair,
	endpoint endpoint.EndpointType,
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
func (h *Handler) AddModelsEndpoint(model interface{}, endpoint *endpoint.Endpoint) error {
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
func (h *Handler) ReplaceModelsEndpoint(model interface{}, endpoint *endpoint.Endpoint) error {
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
