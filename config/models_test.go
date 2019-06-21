package config

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/uni-logger"

	"github.com/neuronlabs/neuron/log"
)

func readConfigFile(t *testing.T, fileName string) {
	t.Helper()

	viper.AddConfigPath("testdata")
	viper.SetConfigName(fileName)

	err := viper.ReadInConfig()
	require.NoErrorf(t, err, "Read config fileName: %s", fileName)
}

// TestModelConfig tests the model config.
func TestModelConfig(t *testing.T) {
	if testing.Verbose() {
		log.SetLevel(unilogger.DEBUG)
	}
	readConfigFile(t, "model")

	cfg := &ModelConfig{}

	require.NoError(t, viper.Unmarshal(cfg))

	assert.Equal(t, "some_model", cfg.Collection)
	assert.Equal(t, "default", cfg.RepositoryName)

}
