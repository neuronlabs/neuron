package jsonapi

import (
	"github.com/kucjac/uni-db"
)

// Repository is an interface that specifies
type Repository interface {
	Create(scope *Scope) *unidb.Error
	Get(scope *Scope) *unidb.Error
	List(scope *Scope) *unidb.Error
	Patch(scope *Scope) *unidb.Error
	Delete(scope *Scope) *unidb.Error
}
