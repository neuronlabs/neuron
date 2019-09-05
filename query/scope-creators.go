package query

import (
	"github.com/neuronlabs/neuron-core/mapping"

	"github.com/neuronlabs/neuron-core/internal"
	"github.com/neuronlabs/neuron-core/internal/safemap"
)

// createModelsRootScope creates scope for given model (mStruct) and
// stores it within the rootScope.includedScopes.
// Used for collection unique root scopes
// (filters, fieldsets etc. for given collection scope)
func (s *Scope) createModelsRootScope(mStruct *mapping.ModelStruct) *Scope {
	rootScope := s.createModelsScope(mStruct)
	rootScope.rootScope.includedScopes[mStruct] = rootScope
	rootScope.includedValues = safemap.New()
	return rootScope
}

// createsModelsScope
func (s *Scope) createModelsScope(mStruct *mapping.ModelStruct) *Scope {
	scope := newScope(mStruct)
	scope.store[internal.ControllerStoreKey] = s.store[internal.ControllerStoreKey]

	if s.rootScope == nil {
		scope.rootScope = s
	} else {
		scope.rootScope = s.rootScope
	}
	return scope
}

// getModelsRootScope returns the scope for given model that is stored within
// the rootScope
func (s *Scope) getModelsRootScope(mStruct *mapping.ModelStruct) (collRootScope *Scope) {
	if s.rootScope == nil {
		// if 's' is root scope and is related to model that is looking for
		if s.mStruct == mStruct {
			return s
		}
		return s.includedScopes[mStruct]
	}
	return s.rootScope.includedScopes[mStruct]
}

// getOrCreateModelsRootScope gets ModelsRootScope and if it is null it creates new.
func (s *Scope) getOrCreateModelsRootScope(mStruct *mapping.ModelStruct) *Scope {
	rootScope := s.getModelsRootScope(mStruct)
	if rootScope == nil {
		rootScope = s.createModelsRootScope(mStruct)
	}
	return rootScope
}
