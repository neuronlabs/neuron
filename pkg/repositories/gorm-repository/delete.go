package gormrepo

import (
	"github.com/kucjac/jsonapi"
	"github.com/kucjac/jsonapi/repositories"
	"github.com/kucjac/uni-db"
)

func (g *GORMRepository) Delete(scope *jsonapi.Scope) error {
	if scope.Value == nil {
		scope.NewValueSingle()
	}

	/**

	  DELETE: PREPARE GORM SCOPE

	*/
	gormScope := g.db.NewScope(scope.Value)
	if err := g.buildFilters(gormScope.DB(), gormScope.GetModelStruct(), scope); err != nil {
		return g.converter.Convert(err)
	}

	/**

	  DELETE: HOOK BEFORE DELETE

	*/

	if beforeDeleter, ok := scope.Value.(repositories.HookRepoBeforeDelete); ok {
		if err := beforeDeleter.RepoBeforeDelete(g.db.New(), scope); err != nil {
			return g.converter.Convert(err)
		}
	}

	/**

	  DELETE: GORM SCOPE DELETE RECORD

	*/
	db := gormScope.DB().Delete(scope.GetValueAddress())
	if err := db.Error; err != nil {
		return g.converter.Convert(err)
	}

	if db.RowsAffected == 0 {
		return unidb.ErrNoResult.New()
	}

	/**

	  DELETE: HOOK AFTER DELETE

	*/
	if afterDeleter, ok := scope.Value.(repositories.HookRepoAfterDelete); ok {
		if err := afterDeleter.RepoAfterDelete(g.db.New(), scope); err != nil {
			return g.converter.Convert(err)
		}
	}

	return nil
}
