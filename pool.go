package flexbuf

import (
	"bytes"
	"sync"
)

// pool of byte buffers.
var pool = sync.Pool{
	New: func() interface{} {
		return make([]byte, bytes.MinRead)
	},
}

// zeroOutSlice zeroes out the byte slice.
func zeroOutSlice(b []byte) {
	for i := range b {
		b[i] = 0
	}
}
