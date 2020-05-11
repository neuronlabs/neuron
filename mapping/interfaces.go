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

// StructFielder is the interfaces used for getting the pointer to itself
type StructFielder interface {
	Self() *StructField
}

// NestedStructFielder is the interface used for nested struct fields.
type NestedStructFielder interface {
	StructFielder
	SelfNested() *NestedField
}
