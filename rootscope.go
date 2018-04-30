package jsonapi

// // RootScope is a scope for the query that
// type RootScope struct {
// 	*Scope

// 	currentErrorCount int
// }

// func newRootScope(mStruct *ModelStruct) *RootScope {
// 	scope := newScope(mStruct)
// 	// root scope contains
// 	root := &RootScope{Scope: scope}
// 	root.IncludedScopes = make(map[*ModelStruct]*Scope)
// 	scope.CollectionScopes[mStruct] = scope
// 	return
// }

// // setIncludedFields sets the included fields for given scope
// func (s *RootScope) setIncludedFields() {

// }

// default cases

//
// Fetch Resource
//
// /api/articles/1?include=author
//
// build scope for articles
// build subscope for author
// add 'author' scope to articles
// add
// set primary filter equal to '1'

//
// Fetch relationship
//
// /api/articles/1/author?include=posts&fields[people]=name
// &&
// /api/arcticles/1/relationships/author?include=posts&fields[people]=name
//
// build scope for articles with id 1
// build subscope for relationship author
// subscope author contains all subscopes for queries
// i.e. subscope for 'posts' from include

// from repository get on
