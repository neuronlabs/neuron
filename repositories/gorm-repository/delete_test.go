package gormrepo

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDeletes(t *testing.T) {
	tests := map[string]func(*testing.T){
		"Simple": func(t *testing.T) {
			models := []interface{}{&Simple{}}
			c, err := prepareJSONAPI(models...)
			require.NoError(t, err)

			repo, err := prepareGORMRepo(models...)
			require.NoError(t, err)

			modelToTest := &Simple{ID: 1, Attr1: "First", Attr2: 2}
			require.NoError(t, repo.db.Create(modelToTest).Error)

			scope, err := c.NewScope(&Simple{})
			require.NoError(t, err)

			scope.SetPrimaryFilters(modelToTest.ID)

			if assert.NoError(t, repo.Delete(scope)) {
				deleted := &Simple{}
				assert.Error(t, repo.db.Model(modelToTest).First(deleted).Error)
			}
		},
		"RelBelongsTo": func(t *testing.T) {
			models := []interface{}{&Comment{}, &Post{}}
			c, err := prepareJSONAPI(models...)
			require.NoError(t, err)

			repo, err := prepareGORMRepo(models...)
			require.NoError(t, err)

			post := &Post{ID: 1, Lang: "pl", Title: "Title"}
			require.NoError(t, repo.db.Create(post).Error)

			comment := &Comment{PostID: post.ID}
			require.NoError(t, repo.db.Create(comment).Error)

			scope, err := c.NewScope(&Comment{})
			require.NoError(t, err)

			scope.SetPrimaryFilters(comment.ID)

			if assert.NoError(t, repo.Delete(scope)) {
				deleted := &Comment{}
				assert.Error(t, repo.db.Model(comment).First(deleted).Error)
			}
		},
		"RelHasOne": func(t *testing.T) {

			models := []interface{}{&Human{}, &BodyPart{}}
			c, err := prepareJSONAPI(models...)
			require.NoError(t, err)

			repo, err := prepareGORMRepo(models...)
			require.NoError(t, err)

			human := &Human{ID: 4}

			nose := &BodyPart{ID: 5, HumanID: human.ID}
			require.NoError(t, repo.db.Create(human).Error)
			require.NoError(t, repo.db.Create(nose).Error)

			scope, err := c.NewScope(&Human{})
			require.NoError(t, err)

			scope.SetPrimaryFilters(human.ID)

			if assert.NoError(t, repo.Delete(scope)) {
				deleted := &Human{}
				assert.Error(t, repo.db.Model(human).Find(deleted).Error)

				clearedFK := &BodyPart{}
				assert.NoError(t, repo.db.Model(nose).Find(clearedFK).Error)

			}

		},
		"RelHasMany": func(t *testing.T) {

		},
		"RelMany2Many": func(t *testing.T) {

		},
	}

	for name, testFunc := range tests {
		t.Run(name, testFunc)
	}
}
