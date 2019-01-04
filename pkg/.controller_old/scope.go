package controller

import (
	aerrors "github.com/kucjac/jsonapi/pkg/errors"
	"github.com/kucjac/jsonapi/pkg/gateway/endpoint"
	"github.com/kucjac/jsonapi/pkg/gateway/modelhandler"
	"github.com/kucjac/jsonapi/pkg/query"
)

func (c *Controller) setPrimaryFilterfield(s *query.Scope, value string) (errs []*aerrors.ApiError) {
	_, errs = buildFilterField(s, s.Struct.collectionType, []string{value}, c, s.Struct, c.Flags, annotationID, operatorEqual)
	return
}

func (c *Controller) SetFlags(s *query.Scope, e *endpoint.Endpoint, mh *modelhandler.ModelHandler) {
	for _, f := range scopeCtxFlags {
		s.Flags().SetFirst(f, e.Flags(), mh.Flags(), c.Flags)
	}
}
