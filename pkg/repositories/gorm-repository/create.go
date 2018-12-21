package gormrepo

import (
	"github.com/kucjac/jsonapi"
	"github.com/kucjac/jsonapi/repositories"
)

func (g *GORMRepository) Create(scope *jsonapi.Scope) error {
	g.log().Debug("CREATE BEGIN")
	defer func() { g.log().Debug("CREATE FINISHED") }()

	db := g.NewDB()

	// Set the JSONAPI pointer into the gorm repo
	g.setJScope(scope, db)

	/**

	  CREATE: HOOK BEFORE CREATE

	*/
	if beforeCreate, ok := scope.Value.(repositories.HookRepoBeforeCreate); ok {
		if err := beforeCreate.RepoBeforeCreate(db, scope); err != nil {

			return g.converter.Convert(err)
		}
	}

	/**

	  CREATE: DB CREATE

	*/

	g.getJScope(db.NewScope(scope.Value))

	modelStruct := db.NewScope(scope.Value).GetModelStruct()

	err := db.Create(scope.Value).Error
	if err != nil {
		return g.converter.Convert(err)
	}

	sc := db.NewScope(scope.Value)
	scope.SetPrimaryFilters(sc.PrimaryKeyValue())

	if err := g.patchNonSyncedRelations(scope, modelStruct, db); err != nil {
		return g.converter.Convert(err)
	}

	/**

	  CREATE: HOOK AFTER CREATE

	*/
	if afterCreate, ok := scope.Value.(repositories.HookRepoAfterCreate); ok {
		if err := afterCreate.RepoAfterCreate(db, scope); err != nil {
			return g.converter.Convert(err)
		}
	}

	return nil
}
