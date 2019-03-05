package repositories

import (
	"github.com/kucjac/jsonapi/mapping"
)

// Repository is the interface used for the repositories
type Repository interface {
	RepositoryName() string
	New(m *mapping.ModelStruct) interface{}
}
