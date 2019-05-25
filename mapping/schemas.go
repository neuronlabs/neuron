package mapping

import (
	"github.com/neuronlabs/neuron/config"
	"github.com/neuronlabs/neuron/internal/models"
)

// Schema is the named collection of models. It may encapsulate different model's with the same model with different logic.
type Schema models.Schema

// Models gets the schema defined models
func (s *Schema) Models() (mStructs []*ModelStruct) {
	for _, m := range (*models.Schema)(s).Models() {
		mStructs = append(mStructs, (*ModelStruct)(m))
	}
	return
}

// Config gets the config.Schema
func (s *Schema) Config() *config.Schema {
	return (*models.Schema)(s).Config()
}