package gormrepo

import (
	"github.com/kucjac/jsonapi"
	"github.com/kucjac/jsonapi/repositories"
	"github.com/kucjac/uni-db"
)

func (g *GORMRepository) Get(scope *jsonapi.Scope) error {

	/**

	  GET: PREPARE GORM SCOPE

	*/
	if scope.Value == nil {
		return IErrNoValuesProvided
	}
	gormScope, err := g.buildScopeGet(scope)
	if err != nil {
		errObj := unidb.ErrInternalError.New()
		errObj.Message = err.Error()
		return errObj
	}

	/**

	  GET: GET SCOPE FROM DB

	*/

	err = gormScope.DB().First(scope.Value).Error
	if err != nil {
		return g.converter.Convert(err)
	}

	/**

	  GET: HOOK AFTER READ

	*/
	if hookAfterRead, ok := scope.Value.(repositories.HookRepoAfterRead); ok {
		if err := hookAfterRead.RepoAfterRead(g.db.New(), scope); err != nil {
			return g.converter.Convert(err)
		}
	}

	return nil
}
