package mapping

// Collectioner is the interface used to get the collection name from the provided model.
type Collectioner interface {
	CollectionName() string
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

type NestedStructFielder interface {
	StructFielder
	SelfNested() *NestedField
}
