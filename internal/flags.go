package internal

import (
	"github.com/kucjac/jsonapi/flags"
)

// ScopeCtxFlags - flags settable on scope, endpoint, model handler and controller
var ScopeCtxFlags = []uint{
	flags.AddMetaCountList,
	flags.UseLinks,
	flags.ReturnPatchContent,
}

// ModelCtxFlags - flags settable for model handler and controller
var ModelCtxFlags = []uint{
	flags.AllowClientID,
}

// ControllerContextFlags - flags settable only for the needs of the controller
var ControllerCtxFlags = []uint{
	flags.AllowForeignKeyFilter,
	flags.UseFilterValueLimit,
}
