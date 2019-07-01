package models

// Collectioner is the interface used to get the collection name from the provided model.
type Collectioner interface {
	CollectionName() string
}
