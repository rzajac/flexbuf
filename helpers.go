package flexbuf

// zeroOutSlice zeroes out the byte slice.
func zeroOutSlice(b []byte) {
	for i := range b {
		b[i] = 0
	}
}
