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
