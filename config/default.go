package config

import (
	"gopkg.in/go-playground/validator.v9"
)

var (
	validate = validator.New()

	// Set defaults for the controller
	defaultValues = map[string]interface{}{
		"naming_convention":      "snake",
		"create_validator_alias": "create",
		"patch_validator_alias":  "patch",
		"log_level":              "info",
		"processor":              defaultThreadsafeProcessorConfig(),
		"repositories":           map[string]*Repository{},
		"included_depth_limit":   2,
	}
)

// Default returns default controller configuration.
func Default() *Controller {
	return defaultConfig()
}

// DefaultValues returns default config values in a spf13/viper compatible way.
func DefaultValues() map[string]interface{} {
	valuesCP := make(map[string]interface{}, len(defaultValues))
	for k, v := range defaultValues {
		valuesCP[k] = v
	}
	return valuesCP
}

func defaultConfig() *Controller {
	return &Controller{
		NamingConvention:     "snake",
		CreateValidatorAlias: "create",
		PatchValidatorAlias:  "patch",
		LogLevel:             "info",
		Processor:            ThreadSafeProcessor(),
		Repositories:         map[string]*Repository{},
		IncludedDepthLimit:   2,
	}
}
