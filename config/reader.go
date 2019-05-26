package config

import (
	"github.com/neuronlabs/neuron/log"
	"github.com/spf13/viper"
	"time"
)

var external bool

var (
	defaultConfig *Config
)

// ViperSetDefaults sets the default values for the viper config
func ViperSetDefaults(v *viper.Viper) {
	setDefaults(v)
}

// ReadNamedConfig reads the config with the provided name
func ReadNamedConfig(name string) (*Config, error) {
	return readNamedConfig(name)
}

// ReadConfig reads the config for given path
func ReadConfig() (*Config, error) {
	return readNamedConfig("config")
}
func readNamedConfig(name string) (*Config, error) {

	v := viper.New()
	v.SetConfigName(name)

	v.AddConfigPath(".")
	v.AddConfigPath("configs")

	setDefaults(v)

	err := v.ReadInConfig()
	if err != nil {
		return nil, err
	}

	c := &Config{}
	if err = v.Unmarshal(c); err != nil {
		log.Debugf("Unmarshaling Controller Config failed. %v", err)
		return nil, err
	}

	return c, nil
}

// ReadDefaultConfig reads the default configuration
func ReadDefaultConfig() *Config {
	return readDefaultConfig()
}

func readDefaultConfig() *Config {

	if defaultConfig == nil {

		v := viper.New()
		setDefaults(v)

		c := &Config{}

		if err := v.Unmarshal(c); err != nil {
			log.Debugf("Unmarshaling Config failed: %v", err)
			panic(err)
		}
		defaultConfig = c
	}

	return defaultConfig
}

// ReadDefaultControllerConfig returns the default controller config
func ReadDefaultControllerConfig() *Controller {
	c := readDefaultConfig()
	return c.Controller
}

// ReadDefaultGatewayConfig returns the default gateway configuration
func ReadDefaultGatewayConfig() *Gateway {
	c := readDefaultConfig()
	return c.Gateway
}

// ReadGatewayConfig reads the gateway config from the provided path and for given config name
func ReadGatewayConfig(name, path string) (*Gateway, error) {
	v := viper.New()
	v.AddConfigPath(path)
	v.SetConfigName(name)
	setDefaultGatewayConfig(v, false)

	err := v.ReadInConfig()
	if err != nil {
		return nil, err
	}

	g := &Gateway{}
	if err = v.Unmarshal(g); err != nil {
		log.Debugf("Unmarshaling Controller Config failed. %v", err)
		return nil, err
	}

	return g, nil
}

// ReadControllerConfig reads the config for the controller
func ReadControllerConfig(name, path string) (*Controller, error) {
	v := viper.New()
	v.AddConfigPath(path)
	v.SetConfigName(name)
	setDefaultControllerConfigs(v, false)

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

// Default values
func setDefaults(v *viper.Viper) {
	setDefaultControllerConfigs(v, true)
	setDefaultGatewayConfig(v, true)
}

func setDefaultControllerConfigs(v *viper.Viper, general bool) {
	// Set defaults for the controller
	keys := map[string]interface{}{
		"naming_convention":             "snake",
		"flags.return_links":            true,
		"flags.use_filter_values_limit": true,
		"flags.return_patch_content":    true,
		"create_validator_alias":        "create",
		"patch_validator_alias":         "patch",
		"default_schema":                "api",
		"processor":                     defaultProcessorConfig(),
	}

	for k, value := range keys {
		if general {
			k = "controller." + k
		}
		v.SetDefault(k, value)
	}

}

func setDefaultGatewayConfig(v *viper.Viper, general bool) {
	// Set Default Gateway config values
	keys := map[string]interface{}{
		"port":                               8080,
		"read_timeout":                       time.Second * 10,
		"read_header_timeout":                time.Second * 5,
		"write_timeout":                      time.Second * 10,
		"idle_timeout":                       time.Second * 120,
		"shutdown_timeout":                   time.Second * 10,
		"router.prefix":                      "v1",
		"i18n.supported_languages":           "[]",
		"router.compression_level":           -1,
		"query_builder.error_limit":          5,
		"query_builder.include_nested_limit": 3,
		"query_builder.filter_value_limit":   50,
		"query_builder.processor":            defaultProcessorConfig(),
	}
	for k, value := range keys {
		if general {
			k = "gateway." + k
		}
		v.SetDefault(k, value)
	}
}
