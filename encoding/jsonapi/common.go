package jsonapi

import (
	"github.com/kucjac/jsonapi/internal"
	"github.com/kucjac/jsonapi/internal/controller"
)

var c *controller.Controller = controller.Default()

const (
	MediaType string = internal.MediaType
)
