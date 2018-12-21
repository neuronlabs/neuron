package gormrepo

import (
	"github.com/kucjac/jsonapi"
	"github.com/kucjac/jsonapi/repositories"
	"github.com/kucjac/uni-db"
	"reflect"
)

func (g *GORMRepository) List(scope *jsonapi.Scope) error {
	g.log().Debug("LIST BEGIN")
	defer func() { g.log().Debug("LIST FINISHED") }()
	if scope.Value == nil {
		return IErrNoValuesProvided
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

	if err = g.getListRelationships(db, scope); err != nil {
		return g.converter.Convert(err)
	}
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
