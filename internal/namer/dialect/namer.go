package dialect

import (
	"github.com/neuronlabs/neuron/internal/models"
)

// FieldNamer is a function used for naming the model's struct field
type FieldNamer func(*models.StructField) string
