package sorts

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestOrder tests the order naming
func TestOrder(t *testing.T) {
	assert.Equal(t, "ascending", AscendingOrder.String())
	assert.Equal(t, "descending", DescendingOrder.String())
}
