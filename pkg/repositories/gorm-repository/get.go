package gormrepo

import (
	"github.com/kucjac/jsonapi/pkg/query/scope"
	"github.com/kucjac/uni-db"
)

// Get is the handler for the GORM repository Getter function
func (g *GORMRepository) Get(s *scope.Scope) error {
	g.log().Debug("START GET")
	defer func() { g.log().Debug("FINISHED GET") }()
	/**

	  GET: PREPARE GORM SCOPE

	*/
	if s.Value == nil {
		return IErrNoValuesProvided
	}
	gormScope, err := g.buildScopeGet(s)
	if err != nil {
		errObj := unidb.ErrInternalError.New()
		errObj.Message = err.Error()
		return errObj
	}

	/**

	  GET: GET SCOPE FROM DB

	*/

	err = gormScope.DB().First(s.Value).Error
	if err != nil {
		return g.converter.Convert(err)
	}

	if err := g.getRelationships(gormScope.DB(), s, s.Value); err != nil {
		dbErr, ok := err.(*unidb.Error)
		if !ok {
			dbErr = g.converter.Convert(err)
		}
		return dbErr

	}

	// /**

	//   GET: HOOK AFTER READ

	// */
	// if hookAfterRead, ok := s.Value.(repositories.HookRepoAfterRead); ok {
	// 	if err := hookAfterRead.RepoAfterRead(g.db.New(), s); err != nil {
	// 		return g.converter.Convert(err)
	// 	}
	// }

	return nil
}
