package errors

import (
	"testing"
)

func TestErrorAddMeta(t *testing.T) {
	err := ErrUnsupportedQueryParameter.Copy()
	err.AddMeta("some", "value")

	assertNotNil(t, err.Meta)
}
