package config

import (
	"github.com/kucjac/jsonapi/log"
	"github.com/kucjac/uni-logger"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func readConfigFile(t *testing.T, fileName string) {
	t.Helper()

	viper.AddConfigPath("testdata")
	viper.SetConfigName(fileName)

	err := viper.ReadInConfig()
	require.NoErrorf(t, err, "Read config fileName: %s", fileName)
}

func TestModelConfig(t *testing.T) {
	if testing.Verbose() {
		log.SetLevel(unilogger.DEBUG)
	}
	readConfigFile(t, "model")

	cfg := &ModelConfig{}

	require.NoError(t, viper.Unmarshal(cfg))

	assert.Equal(t, "some_model", cfg.Collection)
	assert.Equal(t, "default", cfg.Repository)
	// Create
	e := cfg.Endpoints.Create

	assert.Contains(t, e.PresetFilters, "[collection][field][operator]")
	flContainer, err := e.Flags.Container()
	if assert.NoError(t, err) {
		v, ok := flContainer.Get(FlReturnPatchContent)
		assert.True(t, ok)
		assert.True(t, v)

		v, ok = flContainer.Get(FlUseLinks)
		assert.True(t, ok)
		assert.False(t, v)

		_, ok = flContainer.Get(FlAddMetaCountList)
		assert.False(t, ok)
	}
	e = cfg.Endpoints.Get
	assert.True(t, e.Forbidden)

	e = cfg.Endpoints.GetRelated

	assert.Contains(t, e.RelatedField.PresetSorts, "-field")
	assert.Contains(t, e.RelatedField.PresetSorts, "other_field")

}
