package repositories

import (
	"github.com/neuronlabs/neuron/mapping"
	// "github.com/neuronlabs/neuron/query/scope"
)

// RepositoryNamer is the interface used for the repositories to implement
// that defines it's name
type RepositoryNamer interface {
	RepositoryName() string
}

// Repository defines the repository
type Repository interface {
	New(model *mapping.ModelStruct) (Repository, error)
	RepositoryNamer
}

// OptionsSetter is the interface used to set the options from the field's StructField
// Used in repositories to prepare custom structures for the repository defined options.
type OptionsSetter interface {
	SetOptions(field *mapping.StructField)
}
