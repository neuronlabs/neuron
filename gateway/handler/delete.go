package handler

import (
	"github.com/kucjac/jsonapi/errors"
	ictrl "github.com/kucjac/jsonapi/internal/controller"
	"github.com/kucjac/jsonapi/internal/models"
	"github.com/kucjac/jsonapi/internal/query"
	"github.com/kucjac/jsonapi/internal/query/filters"
	"github.com/kucjac/jsonapi/log"
	"github.com/kucjac/jsonapi/mapping"
	"github.com/kucjac/jsonapi/query/scope"
	"github.com/kucjac/uni-db"
	"net/http"
)

// HandleDelete handles the delete query for the provided model
func (h *Handler) HandleDelete(m *mapping.ModelStruct) http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		log.Debugf("[DELETE] Begins for model: '%s'", m.Type().String())
		defer func() { log.Debugf("[DELETE] Finished for model: '%s'", m.Type().String()) }()
		s := scope.NewWithModelC(h.c, m, false)

		// Get the ID from the query
		id, err := query.GetID(req.URL, (*models.ModelStruct)(m))
		if err != nil {
			log.Errorf("HandleDelete->GetID for model: '%s' failed. %v", m.Type(), err)
			h.internalError(rw)
			return
		}

		log.Debugf("URL Get ID value: '%s'", id)

		// Set the Primary Field filter for the scope
		f := filters.NewFilter((*models.ModelStruct)(m).PrimaryField())
		errObj := f.SetValues(
			[]string{id},
			filters.OpEqual,
			(*ictrl.Controller)(h.c).QueryBuilder().I18n,
		)
		if errObj != nil {
			log.Debugf("ClientSide Error. Adding primary filter failed. %v", errObj)
			h.marshalErrors(rw, errObj.IntStatus(), errObj)
			return
		}

		log.Debugf("Primary Filter set for model: '%s' with id: '%s'.", m.Type(), id)

		if err := s.Delete(); err != nil {
			if e, ok := err.(*unidb.Error); ok {
				if e.Compare(unidb.ErrNoResult) {
					errObj := errors.ErrResourceNotFound.Copy()
					errObj.Detail = "Provided object not found"

					log.Debugf("Deleting model: '%s' with id:'%s' failed. Provided object is not found or not accesable. %v", m.Type(), id, errObj)
					h.marshalErrors(rw, errObj.IntStatus(), errObj)
					return
				}
			}
			log.Debugf("Deleting model: '%s' with id: '%s' failed. %v", m.Type(), id)
			h.handleDBError(err, rw)
			return
		}

		log.Debugf("Deleting model: '%s' with id: '%s' succeed.", m.Type(), id)
		rw.WriteHeader(http.StatusNoContent)
	})
}
