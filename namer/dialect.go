package namer

import (
	"github.com/neuronlabs/neuron/mapping"
)

// DialectFieldNamer is the namer function
type DialectFieldNamer func(*mapping.StructField) string
