package mock

import (
	"github.com/kucjac/jsonapi/mapping"
	"github.com/kucjac/jsonapi/query/scope"
)

// Repository is the testing testing repository
type Repository struct{}

func (r *Repository) RepositoryName() string {
	return "testing"
}

func (r *Repository) New(m *mapping.ModelStruct) interface{} {
	return r
}

func (r *Repository) Get(s *scope.Scope) error {
	return nil
}

func (r *Repository) Create(s *scope.Scope) error {
	return nil
}

func (r *Repository) List(s *scope.Scope) error {
	return nil
}

func (r *Repository) Delete(s *scope.Scope) error {
	return nil
}

func (r *Repository) Patch(s *scope.Scope) error {
	return nil
}
