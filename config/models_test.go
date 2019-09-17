package config

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestModelConfig tests the model config.
func TestModelConfig(t *testing.T) {
	mc := map[string]interface{}{"collection": "some_model", "repository_name": "default"}
	cfg := &ModelConfig{}
	v := viper.New()

	v.Set("t", mc)
	err := v.UnmarshalKey("t", cfg)
	require.NoError(t, err)

	assert.Equal(t, "some_model", cfg.Collection)
	assert.Equal(t, "default", cfg.RepositoryName)

}
