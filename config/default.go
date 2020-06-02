package config

// DefaultController returns default controller configuration.
func DefaultController() *Controller {
	return defaultConfig()
}

func defaultConfig() *Controller {
	return &Controller{
		NamingConvention: "snake",
		Services:         map[string]*Service{},
	}
}
