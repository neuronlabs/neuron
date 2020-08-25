package core

// Initializer is an interface used to initialize services, repositories etc.
type Initializer interface {
	Initialize(c *Controller) error
}

// InitializeAll initializes all
func (c *Controller) InitializeAll() (err error) {
	// Initialize all repositories that implements Initializer interface.
	for _, repo := range c.Repositories {
		initializer, ok := repo.(Initializer)
		if ok {
			if err = initializer.Initialize(c); err != nil {
				return err
			}
		}
	}

	// Initialize stores.
	for _, s := range c.Stores {
		initializer, ok := s.(Initializer)
		if ok {
			if err = initializer.Initialize(c); err != nil {
				return err
			}
		}
	}

	// Initialize ss.
	for _, s := range c.FileStores {
		initializer, ok := s.(Initializer)
		if ok {
			if err = initializer.Initialize(c); err != nil {
				return err
			}
		}
	}

	// Initialize authenticators, tokeners and verifiers.
	if c.Authenticator != nil {
		if i, ok := c.Authenticator.(Initializer); ok {
			if err = i.Initialize(c); err != nil {
				return err
			}
		}
	}
	if c.Tokener != nil {
		if i, ok := c.Tokener.(Initializer); ok {
			if err = i.Initialize(c); err != nil {
				return err
			}
		}
	}
	if c.Verifier != nil {
		if i, ok := c.Verifier.(Initializer); ok {
			if err = i.Initialize(c); err != nil {
				return err
			}
		}
	}
	for _, i := range c.Initializers {
		if err = i.Initialize(c); err != nil {
			return err
		}
	}
	return nil
}
