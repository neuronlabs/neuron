package filters

import (
	"github.com/kucjac/jsonapi/pkg/internal/query/filters"
)

var (
	OpEqual        *Operator = (*Operator)(filters.OpEqual)
	OpIn           *Operator = (*Operator)(filters.OpIn)
	OpNotEqual     *Operator = (*Operator)(filters.OpNotEqual)
	OpNotIn        *Operator = (*Operator)(filters.OpNotIn)
	OpGreaterThan  *Operator = (*Operator)(filters.OpGreaterThan)
	OpGreaterEqual *Operator = (*Operator)(filters.OpGreaterEqual)
	OpLessThan     *Operator = (*Operator)(filters.OpLessThan)
	OpLessEqual    *Operator = (*Operator)(filters.OpLessEqual)
	OpContains     *Operator = (*Operator)(filters.OpContains)
	OpStartsWith   *Operator = (*Operator)(filters.OpStartsWith)
	OpEndsWith     *Operator = (*Operator)(filters.OpEndsWith)
	OpIsNull       *Operator = (*Operator)(filters.OpIsNull)
	OpExists       *Operator = (*Operator)(filters.OpExists)
	OpNotExists    *Operator = (*Operator)(filters.OpNotExists)
)
