package filestore

import (
	"github.com/neuronlabs/neuron/mapping"
)

// Model is the interface that defines file model.
type Model interface {
	mapping.Model
	// BucketField returns golang name for the bucket field.
	BucketField() string
	// DirectoryField returns golang name for the directory field.
	DirectoryField() string
	// NameField returns golang name for the file name field.
	NameField() string
	// VersionField returns golang name for the version field.
	VersionField() string
	// ExtensionField returns golang name for the file extension field.
	ExtensionField() string
}
