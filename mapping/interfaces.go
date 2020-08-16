package mapping

// Collectioner is the interface used to get the collection name from the provided model.
type Collectioner interface {
	NeuronCollectionName() string
}

// RepositoryNamer is the interface used for the repositories to implement
// that defines it's name
type RepositoryNamer interface {
	RepositoryName() string
}

// DatabaseNamer is the interface used for defining model's database name - 'Table', 'Collection'.
type DatabaseNamer interface {
	DatabaseName() string
}

// DatabaseSchemaNamer is the interface that defines the optional database schema name for the model.
type DatabaseSchemaNamer interface {
	DatabaseSchemaName() string
}

// Internal interfaces.

// structFielder is the interfaces used for getting the pointer to itself
type structFielder interface {
	Self() *StructField
}

// nestedStructFielder is the interface used for nested struct fields.
type nestedStructFielder interface {
	structFielder
	SelfNested() *NestedField
}
