package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestControllerMapRepositories tests the map repository config function.
func TestControllerMapRepositories(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		c := &Controller{
			Models: map[string]*ModelConfig{
				"model1": &ModelConfig{
					RepositoryName: "testRepoName",
				},
			},
			Repositories: map[string]*Repository{
				"testRepoName": &Repository{DriverName: "testDriver"},
			},
		}

		err := c.MapRepositories()
		require.NoError(t, err)

		model := c.Models["model1"]
		if assert.NotNil(t, model.Repository) {
			assert.Equal(t, "testDriver", model.Repository.DriverName)
		}
	})

	t.Run("DefaultRepository", func(t *testing.T) {
		c := &Controller{
			DefaultRepositoryName: "testRepoName",
			Models: map[string]*ModelConfig{
				"model1": &ModelConfig{},
			},
			Repositories: map[string]*Repository{
				"testRepoName": &Repository{DriverName: "testDriver"},
			},
		}

		err := c.MapRepositories()
		require.NoError(t, err)

		model := c.Models["model1"]
		if assert.NotNil(t, model.Repository) {
			assert.Equal(t, "testDriver", model.Repository.DriverName)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		c := &Controller{
			Models: map[string]*ModelConfig{
				"model1": &ModelConfig{
					RepositoryName: "testRepoName",
				},
			},
			Repositories: map[string]*Repository{
				"differentName": &Repository{DriverName: "testDriver"},
			},
		}

		err := c.MapRepositories()
		require.Error(t, err)
	})
}
