package scope

import (
	scope "github.com/kucjac/jsonapi/pkg/internal/query/scope"
	"github.com/kucjac/jsonapi/pkg/log"
)

var (
	getIncluded Process = Process{
		Name: "whiz:get_included",
		Func: getIncludedFunc,
	}

	setBelongsToRelationships Process = Process{
		Name: "whiz:set_belongs_to_relationships",
		Func: setBelongsToRelationshipsFunc,
	}

	getForeignRelationships Process = Process{
		Name: "whiz:get_foreign_relationships",
		Func: getForeignRelationshipsFunc,
	}
)

// processGetIncluded gets the included fields for the
func getIncludedFunc(s *Scope) error {
	iScope := (*scope.Scope)(s)

	if iScope.IsRoot() && len(iScope.IncludedScopes()) == 0 {
		return nil
	}

	if err := iScope.SetCollectionValues(); err != nil {
		log.Debugf("SetCollectionValues for model: '%v' failed. Err: %v", s.Struct().Collection(), err)
		return err
	}

	for iScope.NextIncludedField() {
		includedField, err := iScope.CurrentIncludedField()
		if err != nil {
			return err
		}

		missing, err := includedField.GetMissingPrimaries()
		if err != nil {
			log.Debugf("Model: %v, includedField: '%s', GetMissingPrimaries failed: %v", s.Struct().Collection(), includedField.Name(), err)
			return err
		}

		if len(missing) > 0 {
			includedField.Scope.SetIDFilters(missing...)
			includedField.Scope.NewValueMany()

			if err := DefaultQueryProcessor.doList((*Scope)(includedField.Scope)); err != nil {
				log.Debugf("Model: %v, includedField '%s' Scope.List failed. %v", s.Struct().Collection(), includedField.Name(), err)
				return err
			}
		}
	}
	iScope.ResetIncludedField()
	return nil
}

func getForeignRelationshipsFunc(s *Scope) error {
	return nil
}

func setBelongsToRelationshipsFunc(s *Scope) error {
	return nil
}
