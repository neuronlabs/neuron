package repositories

import (
	"github.com/kucjac/jsonapi"
	"reflect"
)

/**

READ

*/

type HookRepoAfterRead interface {
	RepoAfterRead(db interface{}, scope *jsonapi.Scope) error
}

func ImplementsHookAfterRead(scope *jsonapi.Scope) bool {
	t := reflect.New(scope.Struct.GetType()).Type()
	return t.Implements(reflect.TypeOf(((*HookRepoAfterRead)(nil))).Elem())
}

/**

CREATE

*/
type HookRepoBeforeCreate interface {
	RepoBeforeCreate(db interface{}, scope *jsonapi.Scope) error
}

func ImplementsHookBeforeCreate(scope *jsonapi.Scope) bool {
	t := reflect.New(scope.Struct.GetType()).Type()
	return t.Implements(reflect.TypeOf(((*HookRepoBeforeCreate)(nil))).Elem())
}

type HookRepoAfterCreate interface {
	RepoAfterCreate(db interface{}, scope *jsonapi.Scope) error
}

func ImplementsHookAfterCreate(scope *jsonapi.Scope) bool {
	t := reflect.New(scope.Struct.GetType()).Type()
	return t.Implements(reflect.TypeOf(((*HookRepoAfterCreate)(nil))).Elem())
}

/**

PATCH

*/

type HookRepoBeforePatch interface {
	RepoBeforePatch(db interface{}, scope *jsonapi.Scope) error
}

func ImplementsHookBeforePatch(scope *jsonapi.Scope) bool {
	t := reflect.New(scope.Struct.GetType()).Type()
	return t.Implements(reflect.TypeOf(((*HookRepoBeforePatch)(nil))).Elem())
}

type HookRepoAfterPatch interface {
	RepoAfterPatch(db interface{}, scope *jsonapi.Scope) error
}

func ImplementsHookAfterPatch(scope *jsonapi.Scope) bool {
	t := reflect.New(scope.Struct.GetType()).Type()
	return t.Implements(reflect.TypeOf(((*HookRepoAfterPatch)(nil))).Elem())
}

/**

DELETE

*/

type HookRepoBeforeDelete interface {
	RepoBeforeDelete(db interface{}, scope *jsonapi.Scope) error
}

func ImplementsHookBeforeDelete(scope *jsonapi.Scope) bool {
	t := reflect.New(scope.Struct.GetType()).Type()
	return t.Implements(reflect.TypeOf(((*HookRepoBeforeDelete)(nil))).Elem())
}

type HookRepoAfterDelete interface {
	RepoAfterDelete(db interface{}, scope *jsonapi.Scope) error
}

func ImplementsHookAfterDelete(scope *jsonapi.Scope) bool {
	t := reflect.New(scope.Struct.GetType()).Type()
	return t.Implements(reflect.TypeOf(((*HookRepoAfterDelete)(nil))).Elem())
}
