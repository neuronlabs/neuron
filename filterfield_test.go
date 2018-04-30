package jsonapi

import (
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
	assertEqual(t, annotationGreaterThan, OpGreaterThan.String())
	assertTrue(t, len(FilterOperator(666).String()) == len("unknown operator"))
}

func TestFilterSetValues(t *testing.T) {
	clearMap()
	scope := getBlogScope()
	errs := scope.buildIncludedScopes("posts")

	assertEmpty(t, errs)
	var f *FilterField
	f, errs = scope.newFilterScope("blogs", []string{"1"}, scope.Struct, "id", "eq")

	assertEmpty(t, errs)
	assertNotNil(t, f)
	errs = f.setValues("blogs", []string{"1"}, FilterOperator(666))
	assertNotEmpty(t, errs)

	f, errs = scope.newFilterScope("blogs", []string{"1"}, scope.Struct, "view_count", "startswith")
	assertNotEmpty(t, errs)

	f, errs = scope.newFilterScope("blogs", []string{"invalid"}, scope.Struct, "view_count", "startswith")
	assertNotEmpty(t, errs)

	clearMap()

	type modelWithStringID struct {
		ID string `jsonapi:"primary,stringers"`
	}
	assertNil(t, c.PrecomputeModels(&modelWithStringID{}))
	mStruct := c.MustGetModelStruct(&modelWithStringID{})
	scope = newRootScope(mStruct, false)
	_, errs = scope.newFilterScope("stringers", []string{"kkk"}, mStruct, "id", "startswith")
	assertNotEmpty(t, errs)

}
