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
			t.Log(err)
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
	t.Log(OpGreaterThan.String())

	assertEqual(t, annotationGreaterThan, OpGreaterThan.String())
	assertTrue(t, len(FilterOperator(666).String()) == len("unknown operator"))
}

func TestFilterSetValues(t *testing.T) {
	clearMap()
	scope := getBlogScope()
	scope.collectionScope = scope
	errs := scope.buildIncludeList("posts")

	assertEmpty(t, errs)
	var f *FilterField
	f, errs = scope.buildFilterfield("blogs", []string{"1"}, scope.Struct, "id", "eq")

	assertEmpty(t, errs)
	assertNotNil(t, f)
	errs = f.setValues("blogs", []string{"1"}, FilterOperator(666))
	assertNotEmpty(t, errs)

	f, errs = scope.buildFilterfield("blogs", []string{"1"}, scope.Struct, "view_count", "startswith")
	assertNotEmpty(t, errs)

	f, errs = scope.buildFilterfield("blogs", []string{"invalid"}, scope.Struct, "view_count", "startswith")
	assertNotEmpty(t, errs)

	clearMap()

	type modelWithStringID struct {
		ID string `jsonapi:"type=primary"`
	}
	assert.NoError(t, c.PrecomputeModels(&modelWithStringID{}))
	mStruct := c.MustGetModelStruct(&modelWithStringID{})
	scope = newScope(mStruct)
	_, errs = scope.buildFilterfield("model_with_string_id", []string{"kkk"}, mStruct, "id", "startswith")
	assertNotEmpty(t, errs)

}
