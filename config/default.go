package config

var (
	// Set defaults for the controller
	defaultValues = map[string]interface{}{
		"naming_convention":    "snake",
		"log_level":            "info",
		"processor":            defaultThreadsafeProcessorConfig(),
		"repositories":         map[string]*Repository{},
		"included_depth_limit": 2,
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
		NamingConvention:   "snake",
		Processor:          ThreadSafeProcessor(),
		Repositories:       map[string]*Repository{},
		IncludedDepthLimit: 2,
	}
}
