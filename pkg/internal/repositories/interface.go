package repositories

import (
	"github.com/kucjac/jsonapi/pkg/mapping"
)

type Repository interface {
	RepositoryName() string
	New(m *mapping.ModelStruct) interface{}
}
