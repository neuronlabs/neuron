package config

import (
	"github.com/kucjac/jsonapi/pkg/log"
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
func ReadDefaultControllerConfig() *ControllerConfig {
	c := readDefaultConfig()
	return c.Controller
}

// ReadDefaultGatewayConfig returns the default gateway configuration
func ReadDefaultGatewayConfig() *GatewayConfig {
	c := readDefaultConfig()
	return c.Gateway
}

func ReadGatewayConfig(name, path string) (*GatewayConfig, error) {
	v := viper.New()
	v.AddConfigPath(path)
	v.SetConfigName(name)

	err := v.ReadInConfig()
	if err != nil {
		return nil, err
	}

	g := &GatewayConfig{}
	if err = v.Unmarshal(g); err != nil {
		log.Debugf("Unmarshaling Controller Config failed. %v", err)
		return nil, err
	}

	return g, nil
}

// ReadControllerConfig reads the config for the controller
func ReadControllerConfig(name, path string) (*ControllerConfig, error) {
	v := viper.New()
	v.AddConfigPath(path)
	v.SetConfigName(name)

	err := v.ReadInConfig()
	if err != nil {
		return nil, err
	}

	c := &ControllerConfig{}
	if err = v.Unmarshal(c); err != nil {
		log.Debugf("Unmarshaling Controller Config failed. %v", err)
		return nil, err
	}

	return c, nil
}

// Default values
func setDefaults(v *viper.Viper) {
	setDefaultControllerConfigs(v)
	setDefaultGatewayConfig(v)
}

func setDefaultControllerConfigs(v *viper.Viper) {
	// Set defaults for the controller
	v.SetDefault("controller.naming_convention", "snake")
	v.SetDefault("controller.builder.error_limit", 5)
	v.SetDefault("controller.builder.include_nested_limit", 3)
	v.SetDefault("controller.builder.filter_value_limit", 50)
	v.SetDefault("controller.flags.return_links", true)
	v.SetDefault("controller.flags.use_filter_values_limit", true)
	v.SetDefault("controller.flags.return_patch_content", true)
	v.SetDefault("controller.create_validator_alias", "create")
	v.SetDefault("controller.patch_validator_alias", "patch")
	v.SetDefault("controller.default_schema", "api")
}

func setDefaultGatewayConfig(v *viper.Viper) {
	// Set Default Gateway config values
	v.SetDefault("gateway.port", 8080)
	v.SetDefault("gateway.read_timeout", time.Second*10)
	v.SetDefault("gateway.read_header_timeout", time.Second*5)
	v.SetDefault("gateway.write_timeout", time.Second*10)
	v.SetDefault("gateway.idle_timeout", time.Second*120)
	v.SetDefault("gateway.shutdown_timeout", time.Second*10)

	v.SetDefault("gateway.router.prefix", "v1")
}
