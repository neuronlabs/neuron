package config

import (
	"github.com/spf13/viper"
	"gopkg.in/go-playground/validator.v9"

	"github.com/neuronlabs/neuron-core/log"
)

var (
	external      bool
	defaultConfig *Controller
	validate      = validator.New()
)

// ViperSetDefaults sets the default values for the viper config.
func ViperSetDefaults(v *viper.Viper) {
	setDefaults(v)
}

// ReadNamedConfig reads the config with the provided name.
func ReadNamedConfig(name string) (*Controller, error) {
	return readNamedConfig(name)
}

// ReadConfig reads the config for given path.
func ReadConfig() (*Controller, error) {
	return readNamedConfig("config")
}

// ReadDefaultConfig reads the default configuration.
func ReadDefaultConfig() *Controller {
	return readDefaultConfig()
}

func readNamedConfig(name string) (*Controller, error) {
	v := viper.New()
	v.SetConfigName(name)

	v.AddConfigPath(".")
	v.AddConfigPath("configs")

	setDefaults(v)

	err := v.ReadInConfig()
	if err != nil {
		return nil, err
	}

	c := &Controller{}
	if err = v.Unmarshal(c); err != nil {
		log.Debugf("Unmarshaling Controller Config failed. %v", err)
		return nil, err
	}

	return c, nil
}

// ReadDefaultControllerConfig returns the default controller config.
func ReadDefaultControllerConfig() *Controller {
	return readDefaultConfig()
}

// ReadControllerConfig reads the config for the controller.
func ReadControllerConfig(name, path string) (*Controller, error) {
	v := viper.New()
	v.AddConfigPath(path)
	v.SetConfigName(name)
	setDefaultControllerConfigs(v)

	err := v.ReadInConfig()
	if err != nil {
		return nil, err
	}

	c := &Controller{}
	if err = v.Unmarshal(c); err != nil {
		log.Debugf("Unmarshaling Controller Config failed. %v", err)
		return nil, err
	}

	return c, nil
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
		"processor":              DefaultProcessorConfig(),
		"repositories":           map[string]*Repository{},
		"included_depth_limit":   2,
	}

	for k, value := range keys {
		v.SetDefault(k, value)
	}
}
