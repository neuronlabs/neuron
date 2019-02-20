package jsonapi

import (
	"github.com/kucjac/jsonapi/pkg/internal"
	"github.com/kucjac/jsonapi/pkg/internal/controller"
)

var c *controller.Controller = controller.Default()

const (
	MediaType string = internal.MediaType
)
