package gormrepo

import (
	iscope "github.com/kucjac/jsonapi/internal/query/scope"
	"github.com/kucjac/jsonapi/query/scope"

	"github.com/kucjac/uni-db"
)

func (g *GORMRepository) List(s *scope.Scope) error {
	g.log().Debug("LIST BEGIN")
	defer func() { g.log().Debug("LIST FINISHED") }()
	if s.Value == nil {
		(*iscope.Scope)(s).NewValueMany()
	}

	/**

	  LIST: BUILD SCOPE LIST

	*/
	g.log().Debugf("buildScopeList")
	gormScope, err := g.buildScopeList(s)
	if err != nil {
		errObj := unidb.ErrInternalError.New()
		errObj.Message = err.Error()
		return errObj
	}

	db := gormScope.DB()

	/**

	  LIST: GET FROM DB

	*/
	g.log().Debugf("db.Find begins")
	err = db.Find(s.Value).Error
	if err != nil {
		g.log().Debugf("%T", s.Value)
		return g.converter.Convert(err)
	}

	g.log().Debugf("getListRelationships begins")
	if err = g.getListRelationships(db, s); err != nil {
		return g.converter.Convert(err)
	}
	/**

	  LIST: HOOK AFTER READ

	*/

	// if repositories.ImplementsHookAfterRead(s) {
	// 	v := reflect.ValueOf(s.Value)
	// 	for i := 0; i < v.Len(); i++ {
	// 		single := v.Index(i).Interface()

	// 		HookAfterRead, ok := single.(repositories.HookRepoAfterRead)
	// 		if ok {
	// 			if err := HookAfterRead.RepoAfterRead(g.db.New(), s); err != nil {
	// 				return g.converter.Convert(err)
	// 			}
	// 		}
	// 		v.Index(i).Set(reflect.ValueOf(single))
	// 	}
	// 	s.Value = v.Interface()
	// }

	return nil
}
