package preset

import (
	"fmt"
	"github.com/neuronlabs/neuron/internal/query/filters"
	"github.com/neuronlabs/neuron/internal/query/scope"
)

// Scope is a struct used by presetting / prechecking given model.
type Scope struct {
	scope  *scope.Scope
	filter *filters.FilterField
	Key    interface{}

	rawPreset string
	rawFilter string
	Error     error
}

// Scope returns preset pair scope
func (p *Scope) Scope() *scope.Scope {
	return p.scope
}

// Filter returns preset pair filter
func (p *Scope) Filter() *filters.FilterField {
	return p.filter
}

func (p *Scope) String() string {
	return fmt.Sprintf("Preset: %s, Filter: %s", p.rawPreset, p.rawFilter)
}

func (p *Scope) GetPair() (*scope.Scope, *filters.FilterField) {
	s := scope.CopyScope(p.scope, nil, true)

	filter := filters.CopyFilter(p.filter)
	return s, filter
}

// WithKey sets the key value for the given preset pair.
// Used to get value from contexts that matches given pair.
func (p *Scope) WithKey(key interface{}) *Scope {
	p.Key = key
	return p
}

// ErrOnFail returns provided error if the given scope fails
func (p *Scope) ErrOnFail(err error) *Scope {
	p.Error = err
	return p
}

func (p *Scope) GetStringKey() string {
	if p.Key == nil {
		return ""
	}
	strKey, ok := p.Key.(string)
	if !ok {
		return ""
	}
	return strKey
}

// Filter is used to point and filter the target scope with preset values.
// It is a FilterField with a Key field.
// The Key field is used to retrieve the value for given filter from the context.
type Filter struct {
	filter *filters.FilterField
	Key    interface{}
}

// Filter returns preset filter's filter field.
func (p *Filter) Filter() *filters.FilterField {
	return p.filter
}

// WithKey sets the key value for the given preset filter.
// Used to get value from the context.
func (p *Filter) WithKey(key interface{}) *Filter {
	p.Key = key
	return p
}

// GetStringKey gets the key for given preset filter in the string form.
func (p *Filter) GetStringKey() string {
	if p.Key == nil {
		return ""
	}
	strKey, ok := p.Key.(string)
	if !ok {
		return ""
	}
	return strKey
}
