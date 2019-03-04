package repositories

import (
	"github.com/kucjac/jsonapi/mapping"
)

type Repository interface {
	RepositoryName() string
	New(m *mapping.ModelStruct) interface{}
}
