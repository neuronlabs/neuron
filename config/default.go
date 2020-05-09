package config

var (
	// AddModel defaults for the controller
	defaultValues = map[string]interface{}{
		"naming_convention": "snake",
		"repositories":      map[string]*Repository{},
	}
)

// DefaultController returns default controller configuration.
func DefaultController() *Controller {
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
		NamingConvention: "snake",
		Repositories:     map[string]*Repository{},
	}
}
