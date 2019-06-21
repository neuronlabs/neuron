package models

// Collectioner is the interface used to get the collection name from the provided model.
type Collectioner interface {
	CollectionName() string
}

// SchemaNamer is the interface used to get the schema name from the model.
type SchemaNamer interface {
	SchemaName() string
}
