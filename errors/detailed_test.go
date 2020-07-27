package errors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestDetailedError tests detailed error functions.
func TestDetailedError(t *testing.T) {
	resetContainer()

	message := "some testing message"
	first := NewDet(ClInvalidIndex, message)
	second := NewDetf(ClInvalidIndex, "formatted: '%d'", 2)

	assert.Equal(t, "some testing message", first.Error())
	assert.Equal(t, "formatted: '2'", second.Error())

	// check operations
	firstOperation := "github.com/neuronlabs/neuron/errors.TestDetailedError#detailed_test.go:14"
	secondOperation := "github.com/neuronlabs/neuron/errors.TestDetailedError#detailed_test.go:15"
	assert.Equal(t, firstOperation, first.Operation)
	assert.Equal(t, secondOperation, second.Operation)

	assert.NotEqual(t, first.ID, second.ID)

	assert.Equal(t, ClInvalidIndex, first.Class())

	detail := "This is detail."
	first.WithDetail(detail)
	assert.Equal(t, detail, first.Details)

	second.WithDetailf("This is %dnd detail.", 2)
	assert.Equal(t, "This is 2nd detail.", second.Details)

	second.Details = ""
	second.WithDetail("Should be stored.")

	assert.Equal(t, "Should be stored.", second.Details)
}
