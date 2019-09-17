package config

import (
	"github.com/spf13/viper"
	"gopkg.in/go-playground/validator.v9"

	"github.com/neuronlabs/neuron-core/log"
)

var validate = validator.New()

// ViperSetDefaults sets the default values for the viper config.
func ViperSetDefaults(v *viper.Viper) {
	setDefaults(v)
}

// ReadDefaultConfig reads the default configuration.
func ReadDefaultConfig() *Controller {
	return readDefaultConfig()
}

func readDefaultConfig() *Controller {
	v := viper.New()
	setDefaults(v)

	c := &Controller{}
	if err := v.Unmarshal(c); err != nil {
		log.Debugf("Unmarshaling Config failed: %v", err)
		panic(err)
	}
	return c
}

// Default values
func setDefaults(v *viper.Viper) {
	setDefaultControllerConfigs(v)
}

func setDefaultControllerConfigs(v *viper.Viper) {
	// Set defaults for the controller
	keys := map[string]interface{}{
		"naming_convention":      "snake",
		"create_validator_alias": "create",
		"patch_validator_alias":  "patch",
		"log_level":              "info",
		"processor":              defaultThreadsafeProcessorConfig(),
		"repositories":           map[string]*Repository{},
		"included_depth_limit":   2,
	}

	for k, value := range keys {
		v.SetDefault(k, value)
	}
}
