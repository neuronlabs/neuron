package jsonapi

import (
	"github.com/kucjac/uni-db"
)

// Repository is an interface that specifies
type Repository interface {
	Create(scope *Scope) error
	Get(scope *Scope) error
	List(scope *Scope) error
	Patch(scope *Scope) error
	Delete(scope *Scope) error
}
