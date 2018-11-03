package gormrepo

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestList(t *testing.T) {
	tests := map[string]func(*testing.T){
		"AttrOnly": func(t *testing.T) {
			models := []interface{}{&Simple{}}
			c, err := prepareJSONAPI(models...)
			require.NoError(t, err)

			repo, err := prepareGORMRepo(models...)
			require.NoError(t, err)

			modelToTest := &Simple{ID: 1, Attr1: "First", Attr2: 2}
			require.NoError(t, repo.db.Create(modelToTest).Error)

			scope, err := c.NewScope(&Simple{})
			require.NoError(t, err)

			scope.NewValueMany()
			scope.SetPrimaryFilters(modelToTest.ID)

			scope.SetAllFields()

			err = repo.List(scope)
			if assert.NoError(t, err) {
				simple, ok := scope.Value.([]*Simple)
				require.True(t, ok)

				assert.Equal(t, modelToTest.ID, simple[0].ID)
				assert.Equal(t, modelToTest.Attr1, simple[0].Attr1)
				assert.Equal(t, modelToTest.Attr2, simple[0].Attr2)
			}
		},
		"RelationBelongsTo": func(t *testing.T) {
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
			scope.NewValueMany()

			require.NoError(t, scope.SetFields(scope.Struct.GetPrimaryField(), "PostID"))

			scope.SetPrimaryFilters(comment.ID)

			err = repo.List(scope)
			if assert.NoError(t, err) {
				comm, ok := scope.Value.([]*Comment)
				require.True(t, ok)

				assert.Equal(t, comment.PostID, comm[0].PostID)
				assert.Equal(t, comment.ID, comm[0].ID)
			}
		},
		"RelationHasOne": func(t *testing.T) {
			t.Run("Synced", func(t *testing.T) {
				models := []interface{}{&Human{}, &BodyPart{}}
				c, err := prepareJSONAPI(models...)
				require.NoError(t, err)

				repo, err := prepareGORMRepo(models...)
				require.NoError(t, err)

				scope, err := c.NewScope(&Human{})
				require.NoError(t, err)

				human := &Human{ID: 4}

				nose := &BodyPart{ID: 5, HumanID: human.ID}
				require.NoError(t, repo.db.Create(human).Error)
				require.NoError(t, repo.db.Create(nose).Error)
				scope.NewValueMany()
				scope.SetPrimaryFilters(human.ID)

				require.NoError(t, scope.SetFields("Nose"))

				err = repo.List(scope)
				if assert.NoError(t, err) {
					hum, ok := scope.Value.([]*Human)
					require.True(t, ok)

					assert.Zero(t, hum[0].Nose)
				}

			})

			t.Run("NonSynced", func(t *testing.T) {
				t.Run("Exists", func(t *testing.T) {
					models := []interface{}{&Human{}, &BodyPart{}}
					c, err := prepareJSONAPI(models...)
					require.NoError(t, err)

					repo, err := prepareGORMRepo(models...)
					require.NoError(t, err)

					scope, err := c.NewScope(&Human{})
					require.NoError(t, err)

					human := &Human{ID: 4}

					nose := &BodyPart{ID: 5, HumanNonSyncID: human.ID}
					require.NoError(t, repo.db.Create(human).Error)
					require.NoError(t, repo.db.Create(nose).Error)

					scope.NewValueMany()

					scope.SetPrimaryFilters(human.ID)
					require.NoError(t, scope.SetFields("NoseNonSynced"))

					err = repo.List(scope)
					if assert.NoError(t, err) {
						hum, ok := scope.Value.([]*Human)
						require.True(t, ok)

						assert.Equal(t, human.ID, hum[0].ID)
						assert.Zero(t, hum[0].Nose)
						if assert.NotZero(t, hum[0].NoseNonSynced) {
							assert.Equal(t, nose.ID, hum[0].NoseNonSynced.ID)
						}
					}
				})

				t.Run("NotExists", func(t *testing.T) {
					models := []interface{}{&Human{}, &BodyPart{}}
					c, err := prepareJSONAPI(models...)
					require.NoError(t, err)

					repo, err := prepareGORMRepo(models...)
					require.NoError(t, err)

					scope, err := c.NewScope(&Human{})
					require.NoError(t, err)

					human := &Human{ID: 4}

					require.NoError(t, repo.db.Create(human).Error)

					scope.NewValueMany()

					scope.SetPrimaryFilters(human.ID)
					require.NoError(t, scope.SetFields("NoseNonSynced"))

					err = repo.List(scope)
					if assert.NoError(t, err) {
						hum, ok := scope.Value.([]*Human)
						require.True(t, ok)
						assert.Equal(t, human.ID, hum[0].ID)
						assert.Zero(t, human.NoseNonSynced)
					}
				})
			})

		},
		"RelationHasMany": func(t *testing.T) {
			t.Run("Synced", func(t *testing.T) {
				models := []interface{}{&Human{}, &BodyPart{}}
				c, err := prepareJSONAPI(models...)
				require.NoError(t, err)

				repo, err := prepareGORMRepo(models...)
				require.NoError(t, err)

				scope, err := c.NewScope(&Human{})
				require.NoError(t, err)

				human := &Human{ID: 4}

				earID1, earID2 := 5, 6
				ears := []*BodyPart{{ID: earID1, HumanID: human.ID}, {ID: earID2, HumanNonSyncID: human.ID}}
				require.NoError(t, repo.db.Create(human).Error)
				for _, ear := range ears {
					require.NoError(t, repo.db.Create(ear).Error)
				}

				scope.NewValueMany()

				scope.SetPrimaryFilters(human.ID)
				require.NoError(t, scope.SetFields("Ears"))

				err = repo.List(scope)
				if assert.NoError(t, err) {
					hum, ok := scope.Value.([]*Human)
					require.True(t, ok)

					assert.Zero(t, hum[0].Ears)
				}

			})
			t.Run("NonSynced", func(t *testing.T) {
				t.Run("Exists", func(t *testing.T) {
					models := []interface{}{&Human{}, &BodyPart{}}
					c, err := prepareJSONAPI(models...)
					require.NoError(t, err)

					repo, err := prepareGORMRepo(models...)
					require.NoError(t, err)

					scope, err := c.NewScope(&Human{})
					require.NoError(t, err)

					human := &Human{ID: 4}

					earID1, earID2 := 5, 6
					ears := []*BodyPart{{ID: earID1, HumanNonSyncID: human.ID}, {ID: earID2, HumanNonSyncID: human.ID}}
					require.NoError(t, repo.db.Create(human).Error)
					for _, ear := range ears {
						require.NoError(t, repo.db.Create(ear).Error)
					}

					scope.NewValueMany()

					scope.SetPrimaryFilters(human.ID)
					require.NoError(t, scope.SetFields("EarsNonSync"))

					err = repo.List(scope)
					if assert.NoError(t, err) {
						hum, ok := scope.Value.([]*Human)
						require.True(t, ok)

						assert.Equal(t, human.ID, hum[0].ID)
						assert.Zero(t, hum[0].Ears)
						if assert.NotZero(t, hum[0].EarsNonSync) {
							var count int
							for _, earNS := range hum[0].EarsNonSync {
								switch earNS.ID {
								case earID1, earID2:
									count++
								default:
									t.FailNow()
								}
							}
							assert.Equal(t, 2, count)
						}
					}
				})

				t.Run("NotExists", func(t *testing.T) {
					models := []interface{}{&Human{}, &BodyPart{}}
					c, err := prepareJSONAPI(models...)
					require.NoError(t, err)

					repo, err := prepareGORMRepo(models...)
					require.NoError(t, err)

					scope, err := c.NewScope(&Human{})
					require.NoError(t, err)

					human1 := &Human{ID: 4}
					human2 := &Human{ID: 10}

					earID1, earID2 := 5, 6
					ears := []*BodyPart{{ID: earID1}, {ID: earID2}}
					require.NoError(t, repo.db.Create(human1).Error)
					require.NoError(t, repo.db.Create(human2).Error)
					for _, ear := range ears {
						require.NoError(t, repo.db.Create(ear).Error)
					}

					scope.NewValueMany()

					scope.SetPrimaryFilters(human1.ID, human2.ID)
					require.NoError(t, scope.SetFields("EarsNonSync"))

					err = repo.List(scope)
					if assert.NoError(t, err) {
						hum, ok := scope.Value.([]*Human)
						require.True(t, ok)

						assert.Len(t, hum, 2)
						for _, hm := range hum {
							assert.Empty(t, hm.EarsNonSync)
						}
					}
				})

				t.Run("MatchExisting", func(t *testing.T) {
					models := []interface{}{&Human{}, &BodyPart{}}
					c, err := prepareJSONAPI(models...)
					require.NoError(t, err)

					repo, err := prepareGORMRepo(models...)
					require.NoError(t, err)

					scope, err := c.NewScope(&Human{})
					require.NoError(t, err)

					humanID1, humanID2 := 10, 12
					humans := []*Human{{ID: humanID1}, {ID: humanID2}}

					earID1, earID2, earID3 := 5, 6, 10
					ears := []*BodyPart{{ID: earID1, HumanNonSyncID: humanID1}, {ID: earID2, HumanNonSyncID: humanID2}, {ID: earID3, HumanNonSyncID: humanID1}}
					for _, hum := range humans {
						require.NoError(t, repo.db.Create(hum).Error)
					}

					for _, ear := range ears {
						require.NoError(t, repo.db.Create(ear).Error)
					}

					scope.NewValueMany()

					scope.SetPrimaryFilters(humanID1, humanID2)
					require.NoError(t, scope.SetFields("EarsNonSync"))

					err = repo.List(scope)
					if assert.NoError(t, err) {
						humansGet, ok := scope.Value.([]*Human)
						require.True(t, ok)

						for _, hum := range humansGet {
							switch hum.ID {
							case humanID1:
								assert.Len(t, hum.EarsNonSync, 2)
							case humanID2:
								assert.Len(t, hum.EarsNonSync, 1)
							}
						}
					}
				})
			})
		},
		"RelationMany2Many": func(t *testing.T) {
			t.Run("Synced", func(t *testing.T) {
				models := []interface{}{&M2MFirst{}, &M2MSecond{}}
				c, err := prepareJSONAPI(models...)
				require.NoError(t, err)

				repo, err := prepareGORMRepo(models...)
				require.NoError(t, err)

				first := &M2MFirst{}
				require.NoError(t, repo.db.Create(first).Error)
				secondOne := &M2MSecond{}
				secondTwo := &M2MSecond{}

				require.NoError(t, repo.db.Create(secondOne).Error)
				require.NoError(t, repo.db.Create(secondTwo).Error)

				require.NoError(t, repo.db.Model(first).Updates(&M2MFirst{ID: first.ID, SecondsSync: []*M2MSecond{secondOne, secondTwo}}).Error)

				scope, err := c.NewScope(&M2MFirst{})
				require.NoError(t, err)
				scope.NewValueMany()

				require.NoError(t, scope.SetFields("SecondsSync"))

				if assert.NoError(t, repo.List(scope)) {
					st, ok := scope.Value.([]*M2MFirst)
					require.True(t, ok)

					assert.Empty(t, st[0].Seconds)
					assert.Empty(t, st[0].SecondsSync)

				}
			})
			t.Run("NonSynced", func(t *testing.T) {
				models := []interface{}{&M2MFirst{}, &M2MSecond{}}
				c, err := prepareJSONAPI(models...)
				require.NoError(t, err)

				repo, err := prepareGORMRepo(models...)
				require.NoError(t, err)

				first := &M2MFirst{}
				require.NoError(t, repo.db.Create(first).Error)
				secondOne := &M2MSecond{}
				secondTwo := &M2MSecond{}

				require.NoError(t, repo.db.Create(secondOne).Error)
				require.NoError(t, repo.db.Create(secondTwo).Error)

				require.NoError(t, repo.db.Model(first).Updates(&M2MFirst{ID: first.ID, Seconds: []*M2MSecond{secondOne, secondTwo}}).Error)

				scope, err := c.NewScope(&M2MFirst{})
				require.NoError(t, err)
				scope.NewValueMany()

				require.NoError(t, scope.SetFields("Seconds"))

				if assert.NoError(t, repo.List(scope)) {
					st, ok := scope.Value.([]*M2MFirst)
					require.True(t, ok)

					assert.Len(t, st[0].Seconds, 2)
					assert.Empty(t, st[0].SecondsSync)

				}
			})
		},
		"Error": func(t *testing.T) {

		},
	}

	for name, testFunc := range tests {
		t.Run(name, testFunc)
	}
}
