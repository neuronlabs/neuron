package scope

import (
	"github.com/pkg/errors"
)

// common errors used in the scope package
var (
	ErrNilScopeValue     error = errors.New("Scope with nil value provided")
	ErrNoRepositoryFound error = errors.New("No repositories found for model.")
	ErrNoGetterRepoFound error = errors.New("No Getter repository possible for provided model")
	ErrNoListerRepoFound error = errors.New("No Lister repository possible for provided model")
	ErrNoPatcherFound    error = errors.New("The repository doesn't implement Patcher interface")
	ErrNoDeleterFound    error = errors.New("The repository doesn't implement Deleter interface")
)
