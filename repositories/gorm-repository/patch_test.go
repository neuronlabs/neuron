package gormrepo

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPatch(t *testing.T) {
	tests := map[string]func(*testing.T){
		"Attribute": func(t *testing.T) {
			c, err := prepareJSONAPI(&UserGORM{}, &PetGORM{})
			require.NoError(t, err)

			repo, err := prepareGORMRepo(&UserGORM{}, &PetGORM{})
			require.NoError(t, err)
			require.NoError(t, settleUsers(repo.db))

			u := &UserGORM{ID: 1, Name: "Marcin"}

			userBefore := &UserGORM{ID: 1}
			require.NoError(t, db.First(userBefore).Error)

			scope, err := c.NewScope(u)
			require.NoError(t, err)

			scope.NewValueSingle()

			scope.SetPrimaryFilters(1)
			assert.NotEmpty(t, scope.PrimaryFilters)
			value := scope.Value.(*UserGORM)
			*value = *u

			scope.SelectedFields = append(scope.SelectedFields, scope.Struct.GetAttributeField("name"))

			assert.Nil(t, repo.Patch(scope))

			userAfter := &UserGORM{ID: 1}
			require.NoError(t, db.First(userAfter).Error)
			assert.Equal(t, u.Name, userAfter.Name)

			assert.Equal(t, userBefore.Surname, userAfter.Surname)
		},
		"RelationBelongsTo": func(t *testing.T) {
			t.Run("SetNew", func(t *testing.T) {
				models := []interface{}{&Post{}, &Comment{}}
				c, err := prepareJSONAPI(models...)
				require.NoError(t, err)

				repo, err := prepareGORMRepo(models...)
				require.NoError(t, err)

				posts := []*Post{{ID: 5}, {ID: 4}}
				for _, post := range posts {
					require.NoError(t, repo.db.Create(post).Error)
				}

				comment := &Comment{PostID: posts[0].ID}
				require.NoError(t, repo.db.Create(comment).Error)

				t.Logf("Comment ID: %v", comment.ID)

				commToPatch := &Comment{ID: comment.ID, PostID: posts[1].ID}

				scope, err := c.NewScope(&Comment{})
				require.NoError(t, err)

				scope.Value = commToPatch

				scope.SetPrimaryFilters(comment.ID)
				t.Logf("CommentToPatchID: %v", commToPatch.ID)

				scope.AddSelectedFields("PostID")

				err = repo.Patch(scope)
				if assert.NoError(t, err) {
					comment := &Comment{ID: comment.ID}

					err = repo.db.First(comment).Error
					assert.NoError(t, err)

					assert.Equal(t, commToPatch.PostID, comment.PostID)

				}
			})

			t.Run("SetToNull", func(t *testing.T) {
				models := []interface{}{&Post{}, &Comment{}}
				c, err := prepareJSONAPI(models...)
				require.NoError(t, err)

				repo, err := prepareGORMRepo(models...)
				require.NoError(t, err)

				post := &Post{ID: 5}

				require.NoError(t, repo.db.Create(post).Error)

				comment := &Comment{PostID: post.ID}
				require.NoError(t, repo.db.Create(comment).Error)

				t.Logf("Comment ID: %v", comment.ID)

				commToPatch := &Comment{ID: comment.ID, PostID: 0}

				scope, err := c.NewScope(&Comment{})
				require.NoError(t, err)

				scope.Value = commToPatch

				scope.SetPrimaryFilters(comment.ID)
				t.Logf("CommentToPatchID: %v", commToPatch.ID)

				scope.AddSelectedFields("PostID")

				err = repo.Patch(scope)
				if assert.NoError(t, err) {
					commentAfter := &Comment{}

					assert.NoError(t, db.Exec("SELECT * FROM comments WHERE post_id IS NULL").First(commentAfter).Error)

					assert.Equal(t, comment.ID, commentAfter.ID)
				}
			})

		},
		"RelationHasOne": func(t *testing.T) {
			t.Run("Synced", func(t *testing.T) {
				t.Log("While patching the relation with the HasOne Synced field it should not get it from the database.")
				models := []interface{}{&Human{}, &BodyPart{}}
				c, err := prepareJSONAPI(models...)
				require.NoError(t, err)

				repo, err := prepareGORMRepo(models...)
				require.NoError(t, err)

				scope, err := c.NewScope(&Human{})
				require.NoError(t, err)

				// Human HasOne Synced Nose

				human := &Human{ID: 4}

				nose := &BodyPart{ID: 5}
				require.NoError(t, repo.db.Create(human).Error)
				require.NoError(t, repo.db.Create(nose).Error)

				scope.Value = &Human{ID: human.ID, Nose: nose}
				require.NoError(t, scope.AddSelectedFields("Nose"))

				err = repo.Patch(scope)
				if assert.NoError(t, err) {
					assert.NoError(t, repo.db.Model(nose).First(nose).Error)

					assert.Zero(t, nose.HumanID)
				}

			})
			t.Run("NonSynced", func(t *testing.T) {
				t.Run("NonZero", func(t *testing.T) {
					t.Run("RelationExists", func(t *testing.T) {
						models := []interface{}{&Human{}, &BodyPart{}}
						c, err := prepareJSONAPI(models...)
						require.NoError(t, err)

						repo, err := prepareGORMRepo(models...)
						require.NoError(t, err)

						scope, err := c.NewScope(&Human{})
						require.NoError(t, err)

						human := &Human{ID: 4}

						otherHuman := &Human{ID: 6}

						nose := &BodyPart{ID: 5, HumanNonSyncID: otherHuman.ID}
						require.NoError(t, repo.db.Create(human).Error)
						require.NoError(t, repo.db.Create(otherHuman).Error)
						require.NoError(t, repo.db.Create(nose).Error)

						scope.Value = &Human{ID: human.ID, NoseNonSynced: &BodyPart{ID: nose.ID}}
						require.NoError(t, scope.AddSelectedFields("NoseNonSynced"))

						scope.SetPrimaryFilters(human.ID)

						err = repo.Patch(scope)
						if assert.NoError(t, err) {
							assert.NoError(t, repo.db.Model(nose).First(nose).Error)
							assert.Equal(t, human.ID, nose.HumanNonSyncID)
						}
					})

					t.Run("RelationNotExists", func(t *testing.T) {
						models := []interface{}{&Human{}, &BodyPart{}}
						c, err := prepareJSONAPI(models...)
						require.NoError(t, err)

						repo, err := prepareGORMRepo(models...)
						require.NoError(t, err)

						scope, err := c.NewScope(&Human{})
						require.NoError(t, err)

						// Human HasOne Synced Nose

						human := &Human{ID: 4}

						nose := &BodyPart{ID: 5}
						require.NoError(t, repo.db.Create(human).Error)
						require.NoError(t, repo.db.Create(nose).Error)

						scope.Value = &Human{ID: human.ID, NoseNonSynced: nose}
						require.NoError(t, scope.AddSelectedFields("NoseNonSynced"))

						scope.SetPrimaryFilters(human.ID)

						err = repo.Patch(scope)
						if assert.NoError(t, err) {
							assert.NoError(t, repo.db.Model(nose).First(nose).Error)
							assert.Equal(t, human.ID, nose.HumanNonSyncID)
						}
					})
				})

				t.Run("Zero", func(t *testing.T) {
					t.Run("RelationExists", func(t *testing.T) {
						models := []interface{}{&Human{}, &BodyPart{}}
						c, err := prepareJSONAPI(models...)
						require.NoError(t, err)

						repo, err := prepareGORMRepo(models...)
						require.NoError(t, err)

						scope, err := c.NewScope(&Human{})
						require.NoError(t, err)

						// Human HasOne Synced Nose

						human := &Human{ID: 4}

						nose := &BodyPart{ID: 5, HumanNonSyncID: human.ID}
						require.NoError(t, repo.db.Create(human).Error)
						require.NoError(t, repo.db.Create(nose).Error)

						scope.Value = &Human{ID: human.ID, NoseNonSynced: nil}
						require.NoError(t, scope.AddSelectedFields("NoseNonSynced"))

						scope.SetPrimaryFilters(human.ID)

						err = repo.Patch(scope)
						if assert.NoError(t, err) {
							clearNose := &BodyPart{ID: nose.ID}
							assert.NoError(t, repo.db.Model(clearNose).First(clearNose).Error)
							assert.NotEqual(t, human.ID, clearNose.HumanNonSyncID)
						}
					})

					t.Run("RelationNotExists", func(t *testing.T) {
						models := []interface{}{&Human{}, &BodyPart{}}
						c, err := prepareJSONAPI(models...)
						require.NoError(t, err)

						repo, err := prepareGORMRepo(models...)
						require.NoError(t, err)

						scope, err := c.NewScope(&Human{})
						require.NoError(t, err)

						// Human HasOne Synced Nose

						human := &Human{ID: 4}

						nose := &BodyPart{ID: 5}

						require.NoError(t, repo.db.Create(human).Error)
						require.NoError(t, repo.db.Create(nose).Error)

						scope.Value = &Human{ID: human.ID, NoseNonSynced: nil}
						require.NoError(t, scope.AddSelectedFields("NoseNonSynced"))

						scope.SetPrimaryFilters(human.ID)

						err = repo.Patch(scope)
						if assert.NoError(t, err) {
							clearNose := &BodyPart{}
							assert.NoError(t, repo.db.Model(BodyPart{ID: nose.ID}).First(clearNose).Error)
							assert.NotEqual(t, human.ID, clearNose.HumanNonSyncID)
						}
					})

				})

			})
		},
		"RelationHasMany": func(t *testing.T) {
			t.Run("Synced", func(t *testing.T) {
				t.Log("While patching the relation with the HasMany Synced field it should not be patched by gorm.")
				models := []interface{}{&Human{}, &BodyPart{}}
				c, err := prepareJSONAPI(models...)
				require.NoError(t, err)

				repo, err := prepareGORMRepo(models...)
				require.NoError(t, err)

				scope, err := c.NewScope(&Human{})
				require.NoError(t, err)

				// Human HasMany Synced Ears

				human := &Human{ID: 4}

				ears := []*BodyPart{{ID: 5}, {ID: 10}}
				require.NoError(t, repo.db.Create(human).Error)
				for _, ear := range ears {
					require.NoError(t, repo.db.Create(ear).Error)
				}

				scope.Value = &Human{ID: human.ID, Ears: ears}
				require.NoError(t, scope.AddSelectedFields("Ears"))
				scope.SetPrimaryFilters(human.ID)

				if assert.NoError(t, repo.Patch(scope)) {
					ears := []*BodyPart{}
					assert.NoError(t, repo.db.Model(human).Association("Ears").Find(&ears).Error)
					assert.Empty(t, ears)
				}

			})
			t.Run("NonSynced", func(t *testing.T) {
				t.Log("While patching the relation with the HasMany NonSynced field, it should be patched  by gorm.")

				t.Run("NotEmpty", func(t *testing.T) {
					models := []interface{}{&Human{}, &BodyPart{}}
					c, err := prepareJSONAPI(models...)
					require.NoError(t, err)

					repo, err := prepareGORMRepo(models...)
					require.NoError(t, err)

					scope, err := c.NewScope(&Human{})
					require.NoError(t, err)

					// Human HasMany Synced Ears

					human := &Human{ID: 4, EarsNonSync: []*BodyPart{{ID: 14, HumanNonSyncID: 4}}}

					ears := []*BodyPart{{ID: 5}, {ID: 10}}
					require.NoError(t, repo.db.Create(human).Error)
					for _, ear := range ears {
						require.NoError(t, repo.db.Create(ear).Error)
					}

					scope.Value = &Human{ID: human.ID, EarsNonSync: []*BodyPart{ears[0], ears[1]}}
					require.NoError(t, scope.AddSelectedFields("EarsNonSync"))
					scope.SetPrimaryFilters(human.ID)

					if assert.NoError(t, repo.Patch(scope)) {
						earsNonSynced := []*BodyPart{}
						assert.NoError(t, repo.db.Where("human_non_sync_id = ?", human.ID).Find(&earsNonSynced).Error)
						// assert.NoError(t, repo.db.Model(human).Association("EarsNonSync").Find(&earsNonSynced).Error)

						var earsCount int
						for _, earNonSynced := range earsNonSynced {
							for _, ear := range ears {
								if earNonSynced.ID == ear.ID {
									earsCount += 1
								}
								assert.Equal(t, human.ID, earNonSynced.HumanNonSyncID)
							}
						}
						assert.Equal(t, len(ears), earsCount)
					}
				})
				t.Run("Empty", func(t *testing.T) {
					t.Run("RelationExists", func(t *testing.T) {
						models := []interface{}{&Human{}, &BodyPart{}}
						c, err := prepareJSONAPI(models...)
						require.NoError(t, err)

						repo, err := prepareGORMRepo(models...)
						require.NoError(t, err)

						scope, err := c.NewScope(&Human{})
						require.NoError(t, err)

						// Human HasMany Synced Ears

						human := &Human{ID: 4}

						ears := []*BodyPart{{ID: 5, HumanNonSyncID: 4}, {ID: 10, HumanNonSyncID: 4}}
						require.NoError(t, repo.db.Create(human).Error)
						for _, ear := range ears {
							require.NoError(t, repo.db.Create(ear).Error)
						}

						scope.Value = &Human{ID: human.ID}
						require.NoError(t, scope.AddSelectedFields("EarsNonSync"))
						scope.SetPrimaryFilters(human.ID)

						if assert.NoError(t, repo.Patch(scope)) {
							earsNonSynced := []*BodyPart{}
							assert.NoError(t, repo.db.Where("human_non_sync_id = ?", human.ID).Find(&earsNonSynced).Error)

							assert.Equal(t, 0, len(earsNonSynced))
						}
					})
					t.Run("RelationNotExists", func(t *testing.T) {
						models := []interface{}{&Human{}, &BodyPart{}}
						c, err := prepareJSONAPI(models...)
						require.NoError(t, err)

						repo, err := prepareGORMRepo(models...)
						require.NoError(t, err)

						scope, err := c.NewScope(&Human{})
						require.NoError(t, err)

						// Human HasMany Synced Ears

						human := &Human{ID: 4}

						ears := []*BodyPart{{ID: 5}, {ID: 10}}
						require.NoError(t, repo.db.Create(human).Error)
						for _, ear := range ears {
							require.NoError(t, repo.db.Create(ear).Error)
						}

						scope.Value = &Human{ID: human.ID}
						require.NoError(t, scope.AddSelectedFields("EarsNonSync"))
						scope.SetPrimaryFilters(human.ID)

						if assert.NoError(t, repo.Patch(scope)) {
							earsNonSynced := []*BodyPart{}
							assert.NoError(t, repo.db.Where("human_non_sync_id = ?", human.ID).Find(&earsNonSynced).Error)

							assert.Equal(t, 0, len(earsNonSynced))
						}
					})
				})

			})
		},
		"RelationMany2Many": func(t *testing.T) {
			t.Run("Synced", func(t *testing.T) {
				t.Logf("While patching the model with related field of type Many2Many synced, the data should not be patched by gorm.")
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

				require.NoError(t, repo.db.Model(first).Association("Seconds").Append(secondOne).Error)

				scope, err := c.NewScope(&M2MFirst{})
				require.NoError(t, err)

				// The value of the second had changed
				scope.Value = &M2MFirst{ID: first.ID, SecondsSync: []*M2MSecond{secondTwo}}
				scope.AddSelectedFields("SecondsSync")
				scope.SetPrimaryFilters(first.ID)

				err = repo.Patch(scope)
				if assert.NoError(t, err) {
					seconds := []*M2MSecond{}
					assert.NoError(t, db.Model(first).Related(&seconds, "SecondsSync").Error)

					assert.Len(t, seconds, 0)
				}
			})

			t.Run("NonSynced", func(t *testing.T) {
				t.Logf("While patching the model with related field of type Many2Many non-synced, the data should be patched by gorm.")
				t.Run("RelationExists", func(t *testing.T) {
					models := []interface{}{&M2MFirst{}, &M2MSecond{}}
					c, err := prepareJSONAPI(models...)
					require.NoError(t, err)

					repo, err := prepareGORMRepo(models...)
					require.NoError(t, err)

					secondOne := &M2MSecond{}
					secondTwo := &M2MSecond{}

					require.NoError(t, repo.db.Create(secondOne).Error)
					require.NoError(t, repo.db.Create(secondTwo).Error)

					first := &M2MFirst{Seconds: []*M2MSecond{{ID: secondOne.ID}, {ID: secondTwo.ID}}}
					require.NoError(t, repo.db.Create(first).Error)

					require.NoError(t, repo.db.Model(first).Association("Seconds").Append(secondOne).Error)

					scope, err := c.NewScope(&M2MFirst{})
					require.NoError(t, err)

					// The value of the second had changed
					scope.Value = &M2MFirst{ID: first.ID, Seconds: []*M2MSecond{secondTwo}}
					scope.AddSelectedFields("Seconds")
					scope.SetPrimaryFilters(first.ID)

					err = repo.Patch(scope)
					if assert.NoError(t, err) {
						seconds := []*M2MSecond{}
						assert.NoError(t, db.Model(first).Related(&seconds, "Seconds").Error)

						assert.Len(t, seconds, 1)

						assert.NotEqual(t, seconds[0].ID, secondOne.ID)
						assert.Equal(t, seconds[0].ID, secondTwo.ID)
					}
				})

				t.Run("RelationNotExists", func(t *testing.T) {
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

					scope, err := c.NewScope(&M2MFirst{})
					require.NoError(t, err)

					// The value of the second had changed
					scope.Value = &M2MFirst{ID: first.ID, Seconds: []*M2MSecond{secondTwo}}
					scope.AddSelectedFields("Seconds")
					scope.SetPrimaryFilters(first.ID)

					err = repo.Patch(scope)
					if assert.NoError(t, err) {
						seconds := []*M2MSecond{}
						assert.NoError(t, db.Model(first).Related(&seconds, "Seconds").Error)

						if assert.Len(t, seconds, 1) {
							assert.Equal(t, secondTwo.ID, seconds[0].ID)
						}

					}
				})
				t.Run("NoBackReference", func(t *testing.T) {
					models := []interface{}{&UnmappedM2m{}, &BodyPart{}}
					c, err := prepareJSONAPI(models...)
					require.NoError(t, err)

					repo, err := prepareGORMRepo(models...)
					require.NoError(t, err)

					scope, err := c.NewScope(&UnmappedM2m{})
					require.NoError(t, err)

					bp := &BodyPart{ID: 1}
					require.NoError(t, repo.db.Create(bp).Error)

					u := &UnmappedM2m{ID: 2}
					require.NoError(t, repo.db.Create(u).Error)

					scope.Value = &UnmappedM2m{ID: 2, BodyParts: []*BodyPart{{ID: 1}}}
					scope.AddSelectedFields("BodyParts")
					scope.SetPrimaryFilters(2)

					if assert.NoError(t, repo.Patch(scope)) {
						scope.SetAllFields()
						scope.SetPrimaryFilters(2)
						u := &UnmappedM2m{}
						scope.Value = u
						assert.NoError(t, repo.Get(scope))

						assert.Len(t, u.BodyParts, 1)

					}
				})
			})
		},
	}

	for name, testFunc := range tests {
		t.Run(name, testFunc)
	}

}
