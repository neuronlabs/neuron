package controller

import (
	"github.com/kucjac/jsonapi/pkg/query"
)

func (c *Controller) setPrimaryFilterfield(s *query.Scope, value string) (errs []*ErrorObject) {
	_, errs = buildFilterField(s, s.Struct.collectionType, []string{value}, c, s.Struct, c.Flags, annotationID, operatorEqual)
	return
}

func (c *Controller) SetFlags(s *query.Scope, e *Endpoint, mh *ModelHandler) {
	for _, f := range scopeCtxFlags {
		s.Flags().SetFirst(f, e.Flags(), mh.Flags(), c.Flags)
	}
}
