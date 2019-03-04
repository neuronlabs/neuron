package gormrepo

import (
	"github.com/kucjac/jsonapi/query/scope"
	"github.com/kucjac/uni-db"
)

// Delete implements deleter interface for the whiz repositories
func (g *GORMRepository) Delete(s *scope.Scope) error {

	/**

	  DELETE: PREPARE GORM SCOPE

	*/
	gormScope := g.db.NewScope(s.Value)
	if err := g.buildFilters(gormScope.DB(), gormScope.GetModelStruct(), s); err != nil {
		return g.converter.Convert(err)
	}

	// /**

	//   DELETE: HOOK BEFORE DELETE

	// */

	// if beforeDeleter, ok := s.Value.(repositories.HookRepoBeforeDelete); ok {
	// 	if err := beforeDeleter.RepoBeforeDelete(g.db.New(), s); err != nil {
	// 		return g.converter.Convert(err)
	// 	}
	// }

	/**

	  DELETE: GORM SCOPE DELETE RECORD

	*/
	v := s.Value
	db := gormScope.DB().Delete(&v)
	if err := db.Error; err != nil {
		return g.converter.Convert(err)
	}

	if db.RowsAffected == 0 {
		return unidb.ErrNoResult.New()
	}

	// /**

	//   DELETE: HOOK AFTER DELETE

	// */
	// if afterDeleter, ok := s.Value.(repositories.HookRepoAfterDelete); ok {
	// 	if err := afterDeleter.RepoAfterDelete(g.db.New(), s); err != nil {
	// 		return g.converter.Convert(err)
	// 	}
	// }

	return nil
}
