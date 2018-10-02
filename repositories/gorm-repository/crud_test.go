package gormrepo

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGORMPatch(t *testing.T) {
	c, err := prepareJSONAPI(&UserGORM{}, &PetGORM{})
	require.NoError(t, err)

	defer clearDB()
	repo, err := prepareGORMRepo(&UserGORM{}, &PetGORM{})
	require.NoError(t, err)
	require.NoError(t, settleUsers(repo.db))

	repo.db.LogMode(true)
	repo.db.Debug()

	u := &UserGORM{ID: 1, Name: "Marcin"}
	t.Run("attribute", func(t *testing.T) {

		userBefore := &UserGORM{ID: 1}
		require.NoError(t, db.First(userBefore).Error)

		scope, err := c.NewScope(u)
		require.NoError(t, err)

		scope.NewValueSingle()

		scope.SetPrimaryFilters(1)
		assert.NotEmpty(t, scope.PrimaryFilters)
		value := scope.Value.(*UserGORM)
		*value = *u

		scope.UpdatedFields = append(scope.UpdatedFields, scope.Struct.GetAttributeField("name"))

		assert.Nil(t, repo.Patch(scope))

		userAfter := &UserGORM{ID: 1}
		assert.NoError(t, db.First(userAfter).Error)
		assert.Equal(t, u.Name, userAfter.Name)

		assert.Equal(t, userBefore.Surname, userAfter.Surname)
	})

	// Case 2:
	// Patch relationship - add another
	t.Run("add_relationship", func(t *testing.T) {
		scope, err := c.NewScope(u)
		require.NoError(t, err)

		userBefore := &UserGORM{ID: 1}
		require.NoError(t, db.First(userBefore).Error)

		scope.NewValueSingle()
		scope.UpdatedFields = append(scope.UpdatedFields, scope.Struct.GetRelationshipField("pets"))

		v := scope.GetValueAddress().(*UserGORM)
		v.ID = 1
		assert.Equal(t, uint(1), v.ID)
		v.Pets = append(v.Pets, &PetGORM{ID: 1}, &PetGORM{ID: 2})
		assert.Nil(t, repo.Patch(scope))

		pets := []*PetGORM{}
		userAfter := &UserGORM{ID: 1}
		require.NoError(t, db.Model(userAfter).Related(&pets, "Pets").Error)
		if assert.Len(t, pets, 2) {
			assert.Equal(t, uint(1), pets[0].ID)
			assert.Equal(t, uint(2), pets[1].ID)
		}
	})

	// Case 3:
	// Patch relationship - erease relations
	t.Run("clear_relationships", func(t *testing.T) {
		scope, err := c.NewScope(u)
		require.NoError(t, err)

		userBefore := &UserGORM{ID: 1}
		require.NoError(t, db.First(userBefore).Error)

		scope.NewValueSingle()

		scope.UpdatedFields = append(scope.UpdatedFields, scope.Struct.GetRelationshipField("pets"))

		v := scope.GetValueAddress().(*UserGORM)
		v.ID = 1
		v.Pets = []*PetGORM{}

		assert.Nil(t, repo.Patch(scope))

		userAfter := &UserGORM{ID: 1}
		require.NoError(t, db.Preload("Pets").First(userAfter).Error)

		assert.Empty(t, userAfter.Pets)
	})

	// Case 4
	t.Run("delete_single_relationship", func(t *testing.T) {
		scope, err := c.NewScope(&PetGORM{})
		require.NoError(t, err)

		petBefore := &PetGORM{ID: 3}
		require.NoError(t, db.Preload("Owner").First(petBefore).Error)
		assert.NotNil(t, petBefore.Owner)

		scope.NewValueSingle()
		scope.UpdatedFields = append(scope.UpdatedFields, scope.Struct.GetRelationshipField("owner"))

		v := scope.GetValueAddress().(*PetGORM)
		v.ID = 3
		v.Owner = nil

		assert.Nil(t, repo.Patch(scope))

		petAfter := &PetGORM{ID: 3}
		if assert.NoError(t, db.Preload("Owner").First(petAfter).Error) {
			assert.Nil(t, petAfter.Owner)
		}

	})
}
