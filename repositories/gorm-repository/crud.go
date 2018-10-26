package gormrepo

import (
	"github.com/kucjac/jsonapi"
	"github.com/kucjac/jsonapi/repositories"
	"github.com/kucjac/uni-db"
	"reflect"
)

func (g *GORMRepository) Get(scope *jsonapi.Scope) error {

	/**

	  GET: PREPARE GORM SCOPE

	*/
	if scope.Value == nil {
		scope.NewValueSingle()
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

	err = gormScope.DB().First(scope.GetValueAddress()).Error
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

func (g *GORMRepository) List(scope *jsonapi.Scope) error {
	if scope.Value == nil {
		scope.NewValueMany()
	}

	/**

	  LIST: BUILD SCOPE LIST

	*/

	gormScope, err := g.buildScopeList(scope)
	if err != nil {
		errObj := unidb.ErrInternalError.New()
		errObj.Message = err.Error()
		return errObj
	}

	db := gormScope.DB()

	/**

	  LIST: GET FROM DB

	*/

	err = db.Find(scope.GetValueAddress()).Error
	if err != nil {
		return g.converter.Convert(err)
	}
	scope.SetValueFromAddressable()

	/**

	  LIST: HOOK AFTER READ

	*/

	if repositories.ImplementsHookAfterRead(scope) {
		v := reflect.ValueOf(scope.Value)
		for i := 0; i < v.Len(); i++ {
			single := v.Index(i).Interface()

			HookAfterRead, ok := single.(repositories.HookRepoAfterRead)
			if ok {
				if err := HookAfterRead.RepoAfterRead(g.db.New(), scope); err != nil {
					return g.converter.Convert(err)
				}
			}
			v.Index(i).Set(reflect.ValueOf(single))
		}
		scope.Value = v.Interface()
	}

	return nil
}

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
