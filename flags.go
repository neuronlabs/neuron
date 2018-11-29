package jsonapi

import (
	"github.com/kucjac/jsonapi/flags"
)

// Flags contexts order (from top to bottom)
// 	- Controller
//	- Model Handler
//	- Endpoint
//	- Scope

// scopeCtxFlags - flags settable on scope, endpoint, model handler and controller
var scopeCtxFlags []uint = []uint{
	flags.AddMetaCountList,
	flags.UseLinks,
	flags.ReturnPatchContent,
}

// modelCtxFlags - flags settable for model handler and controller
var modelCtxFlags []uint = []uint{
	flags.AllowClientID,
}

// controllerContextFlags - flags settable only for the needs of the controller
var controllerCtxFlags []uint = []uint{
	flags.AllowForeignKeyFilter,
	flags.UseFilterValueLimit,
}
