package gormrepo

import (
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCreate(t *testing.T) {
	tests := map[string]func(*testing.T){
		"NoRelations": func(t *testing.T) {

			models := []interface{}{&Simple{}}

			c, err := prepareJSONAPI(models...)
			require.NoError(t, err)

			repo, err := prepareGORMRepo(models...)
			require.NoError(t, err)

			scope, err := c.NewScope(&Simple{})
			require.NoError(t, err)

			scope.Value = &Simple{Attr1: "some", Attr2: 1}

			err = repo.Create(scope)
			if assert.NoError(t, err) {
				simple, ok := scope.Value.(*Simple)
				require.True(t, ok)
				assert.NotZero(t, simple.ID)
			}
		},
		"BelongsTo": func(t *testing.T) {
			t.Run("Synced", func(t *testing.T) {
				t.Skip()
			})

			t.Run("NoSynced", func(t *testing.T) {
				models := []interface{}{&Comment{}, &Post{}}

				c, err := prepareJSONAPI(models...)
				require.NoError(t, err)

				repo, err := prepareGORMRepo(models...)
				require.NoError(t, err)

				post := &Post{ID: 1, Lang: "pl", Title: "Title"}
				require.NoError(t, repo.db.Create(post).Error)

				scope, err := c.NewScope(models[0])
				require.NoError(t, err)

				scope.Value = &Comment{ID: 2, Body: "Some body", Post: &Post{ID: 1}, PostID: 1}
				scope.AddSelectedFields("id", "body", "post", "post_id")

				err = repo.Create(scope)
				if assert.NoError(t, err) {
					comments := []*Comment{}
					assert.NoError(t, repo.db.Model(post).Related(&comments).Error)

					comment, ok := scope.Value.(*Comment)
					require.True(t, ok)
					if assert.Len(t, comments, 1) {
						relComment := comments[0]
						if assert.NotNil(t, relComment) {
							assert.Equal(t, comment.ID, relComment.ID)
						}
					}
				}
			})

		},
		"HasOne": func(t *testing.T) {
			t.Run("Synced", func(t *testing.T) {
				t.Log("Creating an object with synced relationship of type HasOne should not add related value.")

				models := []interface{}{&Comment{}, &Post{}, &Blog{}, &User{}, &House{}}

				c, err := prepareJSONAPI(models...)
				require.NoError(t, err)

				repo, err := prepareGORMRepo(models...)
				require.NoError(t, err)

				post := &Post{ID: 5}
				require.NoError(t, repo.db.Create(post).Error)

				blog := &Blog{Title: "MyTile", CurrentPost: post}

				scope, err := c.NewScope(blog)
				require.NoError(t, err)

				scope.Value = blog

				scope.AddSelectedFields("title", "current_post")

				err = repo.Create(scope)
				if assert.NoError(t, err) {
					relPost := &Post{}
					err = db.Model(blog).Related(relPost).Error
					if err != gorm.ErrRecordNotFound {
						t.Fail()
						return
					}

					assert.NotZero(t, blog.ID)
				}
			})

			t.Run("NoSynced", func(t *testing.T) {
				t.Log("Creating an object with NonSynced relationship of type HasOne should add related value.")

				models := []interface{}{&Human{}, &BodyPart{}}

				c, err := prepareJSONAPI(models...)
				require.NoError(t, err)

				repo, err := prepareGORMRepo(models...)
				require.NoError(t, err)

				nose := &BodyPart{ID: 5}
				require.NoError(t, repo.db.Create(nose).Error)

				human := &Human{NoseNonSynced: nose}

				scope, err := c.NewScope(human)
				require.NoError(t, err)

				scope.Value = human
				scope.AddSelectedFields("nose_non_synced")

				err = repo.Create(scope)
				if assert.NoError(t, err) {
					noseNonSync := &BodyPart{}

					err = repo.db.Model(scope.Value).Association("NoseNonSynced").Find(noseNonSync).Error
					assert.NoError(t, err)

					assert.Equal(t, nose.ID, noseNonSync.ID)
				}

			})

		},
		"HasMany": func(t *testing.T) {
			t.Run("Synced", func(t *testing.T) {
				models := []interface{}{&Human{}, &BodyPart{}}

				c, err := prepareJSONAPI(models...)
				require.NoError(t, err)

				repo, err := prepareGORMRepo(models...)
				require.NoError(t, err)

				ears := []*BodyPart{{ID: 4}, {ID: 5}}
				for _, ear := range ears {
					c.Log().Debugf("Inserting ear: %v", ear)
					require.NoError(t, repo.db.Create(ear).Error)
				}

				human := &Human{Ears: ears}

				scope, err := c.NewScope(human)
				require.NoError(t, err)

				scope.Value = human

				err = repo.Create(scope)
				if assert.NoError(t, err) {
					earsSynced := []*BodyPart{}
					err = repo.db.Model(human).Association("Ears").Find(&earsSynced).Error
					if assert.NoError(t, err) {
						assert.NotZero(t, human.ID)
						assert.Empty(t, earsSynced)
					}

				}
			})

			t.Run("NoSynced", func(t *testing.T) {
				models := []interface{}{&Human{}, &BodyPart{}}

				c, err := prepareJSONAPI(models...)
				require.NoError(t, err)

				repo, err := prepareGORMRepo(models...)
				require.NoError(t, err)

				ears := []*BodyPart{{ID: 4}, {ID: 5}}
				for _, ear := range ears {
					c.Log().Debugf("Inserting ear: %v", ear)
					require.NoError(t, repo.db.Create(ear).Error)
				}

				human := &Human{EarsNonSync: ears}

				scope, err := c.NewScope(human)
				require.NoError(t, err)

				scope.Value = human
				scope.AddSelectedFields("ears_non_sync")

				err = repo.Create(scope)
				if assert.NoError(t, err) {
					earsNonSync := []*BodyPart{}
					err = repo.db.Model(scope.Value).Association("EarsNonSync").Find(&earsNonSync).Error

					if assert.NoError(t, err) {
						assert.Len(t, earsNonSync, 2)

					}

				}
			})

		},
		"Many2Many": func(t *testing.T) {
			t.Run("Synced", func(t *testing.T) {
				models := []interface{}{&M2MFirst{}, &M2MSecond{}}

				c, err := prepareJSONAPI(models...)
				require.NoError(t, err)

				repo, err := prepareGORMRepo(models...)
				require.NoError(t, err)

				m2mFirst := &M2MFirst{SecondsSync: []*M2MSecond{{ID: 5}}}

				scope, err := c.NewScope(m2mFirst)
				require.NoError(t, err)

				scope.Value = m2mFirst
				scope.AddSelectedFields("seconds_sync")

				err = repo.Create(scope)
				if assert.NoError(t, err) {
					assert.NotZero(t, m2mFirst.ID)

					m2m := M2MFirst{}
					err = db.Model(m2mFirst).First(&m2m).Error
					assert.NoError(t, err)
				}

			})

			t.Run("NoSynced", func(t *testing.T) {
				models := []interface{}{&M2MFirst{}, &M2MSecond{}}

				c, err := prepareJSONAPI(models...)
				require.NoError(t, err)

				repo, err := prepareGORMRepo(models...)
				require.NoError(t, err)

				repo.db.Create(&M2MSecond{ID: 5})
				m2mFirst := &M2MFirst{Seconds: []*M2MSecond{{ID: 5}}}

				scope, err := c.NewScope(m2mFirst)
				require.NoError(t, err)

				scope.Value = m2mFirst
				require.NoError(t, scope.AddSelectedFields("seconds"))

				err = repo.Create(scope)
				if assert.NoError(t, err) {
					assert.NotZero(t, m2mFirst.ID)

					seconds := []*M2MSecond{}
					err = repo.db.Model(m2mFirst).Related(&seconds, "Seconds").Error
					assert.NoError(t, err)

					assert.Len(t, seconds, len(m2mFirst.Seconds))
				}
			})

		},
		"Error": func(t *testing.T) {
			models := []interface{}{&Simple{}}

			c, err := prepareJSONAPI(models...)
			require.NoError(t, err)

			repo, err := prepareGORMRepo(models...)
			require.NoError(t, err)

			scope, err := c.NewScope(&Simple{})
			require.NoError(t, err)
			require.NoError(t, repo.db.Create(&Simple{ID: 1}).Error)

			// create duplicate
			scope.Value = &Simple{ID: 1, Attr1: "some", Attr2: 1}
			scope.AddSelectedFields("id", "attr1", "attr2")

			err = repo.Create(scope)
			assert.Error(t, err)

		},
	}

	for name, testFunc := range tests {
		t.Run(name, testFunc)
	}
}
