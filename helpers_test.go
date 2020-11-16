package flexbuf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_helpers_zeroOutSlice(t *testing.T) {
	// --- Given ---
	data := []byte{0, 1, 2, 3}

	// --- When ---
	zeroOutSlice(data)

	// --- Then ---
	assert.Exactly(t, []byte{0, 0, 0, 0}, data)
}
