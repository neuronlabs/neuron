package query

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSplitBracketParameter(t *testing.T) {
	type stringBool struct {
		Str string
		Val bool
	}

	values := []stringBool{
		{"[some][thing]", true},
		{"[no][closing", false},
		{"no][opening]", false},
		{"]justclosing", false},
		{"[doubleopen[]", false},
		{"[doubleclose]]", false},
	}

	var splitted []string
	var err error
	for _, v := range values {
		splitted, err = splitBracketParameter(v.Str)
		if !v.Val {
			assert.Error(t, err)
			// t.Log(err)
			if err == nil {
				t.Log(v.Str)
			}
		} else {
			assert.Nil(t, err)
			assert.NotEmpty(t, splitted)
		}
	}
}

func TestFilterOperators(t *testing.T) {
	assert.True(t, OpEqual.isBasic())
	assert.False(t, OpEqual.isRangable())
	assert.False(t, OpEqual.isStringOnly())
	assert.True(t, OpContains.isStringOnly())
	// t.Log(OpGreaterThan.String())

	assertEqual(t, operatorGreaterThan, OpGreaterThan.String())
	assert.True(t, len(FilterOperator(666).String()) == len("unknown operator"))
}

func TestFilterSetValues(t *testing.T) {
	clearMap()
	scope := getBlogScope()
	scope.collectionScope = scope
	errs := scope.buildIncludeList("posts")

	assert.Empty(t, errs)
	var f *Field
	f, errs = buildField(scope, "blogs", []string{"1"}, c, scope.Struct, scope.Flags(), "id", "$eq")

	assert.Empty(t, errs)
	assert.NotNil(t, f)
	errs = setFilterValues(f, c, "blogs", []string{"1"}, FilterOperator(666))
	assert.NotEmpty(t, errs)

	f, errs = buildField(scope, "blogs", []string{"1"}, c, scope.Struct, scope.Flags(), "view_count", "$startswith")
	assert.NotEmpty(t, errs)

	f, errs = buildField(scope, "blogs", []string{"invalid"}, c, scope.Struct, scope.Flags(), "view_count", "$startswith")
	assert.NotEmpty(t, errs)

	clearMap()

	type modelWithStringID struct {
		ID string `jsonapi:"type=primary"`
	}
	assert.NoError(t, c.PrecomputeModels(&modelWithStringID{}))
	mStruct := c.MustGetModelStruct(&modelWithStringID{})
	scope = newScope(mStruct)
	_, errs = buildField(scope, "model_with_string_id", []string{"kkk"}, c, mStruct, scope.Flags(), "id", "$startswith")
	assert.NotEmpty(t, errs)

}

func TestMapFilter(t *testing.T) {
	clearMap()

	type MapStruct struct {
		ID  string                 `jsonapi:"type=primary"`
		Map map[string]interface{} `jsonapi:"type=attr"`
	}

	assert.NoError(t, c.PrecomputeModels(&MapStruct{}))

	mStruct := c.MustGetModelStruct(&MapStruct{})

	tests := map[string]func(*testing.T){
		"keyFilter": func(t *testing.T) {
			scope := newScope(mStruct)
			_, errs := buildField(scope, mStruct.collectionType, []string{"someValue"}, c, mStruct, scope.Flags(), "map", "someKey")
			assert.Empty(t, errs)
		},
		"filterExists": func(t *testing.T) {
			scope := newScope(mStruct)
			_, errs := buildField(scope, mStruct.collectionType, []string{}, c, mStruct, scope.Flags(), "map", "somekey", "$exists")
			assert.Empty(t, errs)
		},
		"filterNullKey": func(t *testing.T) {
			scope := newScope(mStruct)
			ff, errs := buildField(scope, mStruct.collectionType, []string{}, c, mStruct, scope.Flags(), "map", "some", "$isnull")
			if assert.Empty(t, errs) {
				if assert.NotEmpty(t, ff.Nested) {
					n := ff.Nested[0]
					assert.Equal(t, "some", n.Key)
					if assert.NotEmpty(t, n.Values) {
						nv := n.Values[0]
						assert.Equal(t, OpIsNull, nv.Operator)
					}
				}
			}
		},
	}

	for name, test := range tests {
		t.Run(name, test)
	}

}
