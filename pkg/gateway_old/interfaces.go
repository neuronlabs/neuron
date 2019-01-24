package gateway

import (
	"github.com/kucjac/jsonapi/pkg/controller"
	"github.com/kucjac/jsonapi/pkg/gateway/endpoint"
	"github.com/kucjac/jsonapi/pkg/query"
)

type ModelPresetGetter interface {
	GetPresetPair(endpoint endpoint.EndpointType, controller *controller.Controller) *query.PresetPair
}

type ModelPrecheckGetter interface {
	GetPrecheckPair(endpoint endpoint.EndpointType, controller *controller.Controller) *query.PresetPair
}
