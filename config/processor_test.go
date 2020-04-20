package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestProcessor tests the processor config.
func TestProcessor(t *testing.T) {
	var p *Processor
	t.Run("ThreadSafe", func(t *testing.T) {
		assert.NotPanics(t, func() { p = ThreadSafeProcessor() })
		if assert.NotNil(t, p) {
			err := p.Validate()
			assert.NoError(t, err)
		}
	})

	t.Run("Concurrent", func(t *testing.T) {
		assert.NotPanics(t, func() { p = ConcurrentProcessor() })
		if assert.NotNil(t, p) {
			err := p.Validate()
			assert.NoError(t, err)
		}
	})
}
