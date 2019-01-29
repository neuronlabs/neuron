package gormrepo

import (
	"github.com/kucjac/jsonapi/pkg/query/filters"
	"github.com/kucjac/jsonapi/pkg/query/scope"
)

func (g *GORMRepository) Create(s *scope.Scope) error {
	g.log().Debug("[GORMREPO][CREATE] begins")
	defer func() { g.log().Debug("[GORMREPO][CREATE] finished") }()

	db := g.NewDB()

	// Set the JSONAPI pointer into the gorm repo
	g.setJScope(s, db)

	// /**

	//   CREATE: HOOK BEFORE CREATE

	// */
	// if beforeCreate, ok := s.Value.(repositories.HookRepoBeforeCreate); ok {
	// 	if err := beforeCreate.RepoBeforeCreate(db, s); err != nil {

	// 		return g.converter.Convert(err)
	// 	}
	// }

	/**

	  CREATE: DB CREATE

	*/

	g.getJScope(db.NewScope(s.Value))

	modelStruct := db.NewScope(s.Value).GetModelStruct()

	err := db.Create(s.Value).Error
	if err != nil {
		return g.converter.Convert(err)
	}

	sc := db.NewScope(s.Value)

	// Add the primary filter
	s.AddFilter(filters.NewFilter(s.Struct().Primary(), filters.OpEqual, sc.PrimaryKeyValue()))

	if err := g.patchNonSyncedRelations(s, modelStruct, db); err != nil {
		return g.converter.Convert(err)
	}

	// /**

	//   CREATE: HOOK AFTER CREATE

	// */
	// if afterCreate, ok := s.Value.(repositories.HookRepoAfterCreate); ok {
	// 	if err := afterCreate.RepoAfterCreate(db, s); err != nil {
	// 		return g.converter.Convert(err)
	// 	}
	// }

	return nil
}
