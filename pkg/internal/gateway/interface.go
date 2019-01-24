package gateway

import (
	"github.com/kucjac/jsonapi/pkg/internal/query/scope"
)

// BeforeCreator is the interface used before create method occurs
type BeforeCreator interface {
	BeforeCreate(s *scope.Scope) error
}
