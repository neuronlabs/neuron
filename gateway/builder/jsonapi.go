package builder

import (
	"github.com/neuronlabs/neuron/config"
	"github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/i18n"
	ictrl "github.com/neuronlabs/neuron/internal/controller"

	"github.com/neuronlabs/neuron/internal/models"
)

// JSONAPI is a struct that is responsible for creating the query scopes
type JSONAPI struct {
	// I18n defines current builder i18n support
	I18n *i18n.Support

	// Config is the configuration for the builder
	Config *config.Builder

	// schemas are the given model schemas
	schemas *models.ModelSchemas
}

// DefaultConfig is the default builder config

// NewJSONAPI creates new query builder
func NewJSONAPI(c *controller.Controller, cfg *config.Builder, i18nSupport *i18n.Support) *JSONAPI {
	modelsSchemas := (*ictrl.Controller)(c).ModelSchemas()
	b := &JSONAPI{schemas: modelsSchemas, I18n: i18nSupport, Config: cfg}

	modelsSchemas.ComputeNestedIncludedCount(cfg.IncludeNestedLimit)
	return b
}
