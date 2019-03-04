package handler

import (
	"github.com/kucjac/jsonapi/mapping"
	"net/http"
)

// HandlePatchRelationship handles the patch relationship API query
func (h *Handler) HandlePatchRelationship(m *mapping.ModelStruct) http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {

	})
}
