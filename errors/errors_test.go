package errors

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestErrorAddMeta(t *testing.T) {
	err := ErrUnsupportedQueryParameter.Copy()
	err.AddMeta("some", "value")

	assert.NotNil(t, err.Meta)
}
