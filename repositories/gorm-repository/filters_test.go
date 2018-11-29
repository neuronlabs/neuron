package gormrepo

import (
	"github.com/kucjac/jsonapi"
	"reflect"
	// "github.com/kucjac/jsonapi"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBuildRelationshipFilter(t *testing.T) {
	defer clearDB()

	c, err := prepareJSONAPI(blogModels...)
	assert.NoError(t, err)

	repo, err := prepareGORMRepo(blogModels...)

	assert.NoError(t, err)
	// repo.db.Debug()
	// repo.db.LogMode(true)

	assert.NoError(t, settleBlogs(repo.db))

	// Case 1:
	// BelongsTo relationship.
	_, req := getHttpPair("GET", "/blogs?filter[blogs][current_post][id][$lt]=3", nil)
	scope, errs, err := c.BuildScopeList(req, &jsonapi.Endpoint{Type: jsonapi.List}, &jsonapi.ModelHandler{ModelType: reflect.TypeOf(Blog{})})
	assert.NoError(t, err)
	assert.Empty(t, errs)

	scope.NewValueMany()

	dbErr := repo.List(scope)
	assert.Nil(t, dbErr)

	// Case 2;
	// HasMany relationship.
	_, req = getHttpPair("GET", "/users?filter[users][blogs][id][$in]=1,3", nil)
	scope, errs, err = c.BuildScopeList(req, &jsonapi.Endpoint{Type: jsonapi.List}, &jsonapi.ModelHandler{ModelType: reflect.TypeOf(User{})})
	assert.NoError(t, err)
	assert.Empty(t, errs)

	scope.NewValueMany()
	dbErr = repo.List(scope)
	assert.Nil(t, dbErr)

	// Case 3:
	// Many2Many relationship
	_, req = getHttpPair("GET", "/users?filter[users][houses][id][$in]=1,2", nil)
	scope, errs, err = c.BuildScopeList(req, &jsonapi.Endpoint{Type: jsonapi.List}, &jsonapi.ModelHandler{ModelType: reflect.TypeOf(User{})})
	assert.NoError(t, err)
	assert.Empty(t, errs)

	scope.NewValueMany()
	dbErr = repo.List(scope)
	assert.Nil(t, dbErr)
}
