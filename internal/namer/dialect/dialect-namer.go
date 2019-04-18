package dialect

import (
	"github.com/neuronlabs/neuron/internal/models"
)

type DialectFieldNamer func(*models.StructField) string
