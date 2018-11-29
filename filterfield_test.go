package jsonapi

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
			assertError(t, err)
			// t.Log(err)
			if err == nil {
				t.Log(v.Str)
			}
		} else {
			assertNil(t, err)
			assertNotEmpty(t, splitted)
		}
	}
}

func TestFilterOperators(t *testing.T) {
	assertTrue(t, OpEqual.isBasic())
	assertFalse(t, OpEqual.isRangable())
	assertFalse(t, OpEqual.isStringOnly())
	assertTrue(t, OpContains.isStringOnly())
	// t.Log(OpGreaterThan.String())

	assertEqual(t, operatorGreaterThan, OpGreaterThan.String())
	assertTrue(t, len(FilterOperator(666).String()) == len("unknown operator"))
}

func TestFilterSetValues(t *testing.T) {
	clearMap()
	scope := getBlogScope()
	scope.collectionScope = scope
	errs := scope.buildIncludeList("posts")

	assertEmpty(t, errs)
	var f *FilterField
	f, errs = buildFilterField(scope, "blogs", []string{"1"}, c, scope.Struct, scope.Flags(), "id", "$eq")

	assertEmpty(t, errs)
	assertNotNil(t, f)
	errs = setFilterValues(f, c, "blogs", []string{"1"}, FilterOperator(666))
	assertNotEmpty(t, errs)

	f, errs = buildFilterField(scope, "blogs", []string{"1"}, c, scope.Struct, scope.Flags(), "view_count", "$startswith")
	assertNotEmpty(t, errs)

	f, errs = buildFilterField(scope, "blogs", []string{"invalid"}, c, scope.Struct, scope.Flags(), "view_count", "$startswith")
	assertNotEmpty(t, errs)

	clearMap()

	type modelWithStringID struct {
		ID string `jsonapi:"type=primary"`
	}
	assert.NoError(t, c.PrecomputeModels(&modelWithStringID{}))
	mStruct := c.MustGetModelStruct(&modelWithStringID{})
	scope = newScope(mStruct)
	_, errs = buildFilterField(scope, "model_with_string_id", []string{"kkk"}, c, mStruct, scope.Flags(), "id", "$startswith")
	assertNotEmpty(t, errs)

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
			_, errs := buildFilterField(scope, mStruct.collectionType, []string{"someValue"}, c, mStruct, scope.Flags(), "map", "someKey")
			assert.Empty(t, errs)
		},
		"filterExists": func(t *testing.T) {
			scope := newScope(mStruct)
			_, errs := buildFilterField(scope, mStruct.collectionType, []string{}, c, mStruct, scope.Flags(), "map", "somekey", "$exists")
			assert.Empty(t, errs)
		},
		"filterNullKey": func(t *testing.T) {
			scope := newScope(mStruct)
			ff, errs := buildFilterField(scope, mStruct.collectionType, []string{}, c, mStruct, scope.Flags(), "map", "some", "$isnull")
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
