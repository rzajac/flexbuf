package flexbuf

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_New_Offset_Negative(t *testing.T) {
	assert.Panics(t, func() {
		New(Offset(-1))
	})
}

func Test_With(t *testing.T) {
	// --- When ---
	buf := With([]byte{0, 1, 2})

	// --- Then ---
	assert.Exactly(t, 0, buf.flag)
	assert.Exactly(t, 0, buf.off)
	assert.Exactly(t, []byte{0, 1, 2}, buf.buf)
}

func Test_With_Offset(t *testing.T) {
	// --- When ---
	buf := With([]byte{0, 1, 2}, Offset(1))

	// --- Then ---
	assert.Exactly(t, 0, buf.flag)
	assert.Exactly(t, 1, buf.off)
	assert.Exactly(t, []byte{0, 1, 2}, buf.buf)
}

func Test_With_Offset_Negative(t *testing.T) {
	assert.Panics(t, func() {
		With([]byte{0, 1, 2}, Offset(-1))
	})
}

func Test_With_Offset_BeyondLen(t *testing.T) {
	assert.Panics(t, func() {
		With([]byte{0, 1, 2}, Offset(5))
	})
}

func Test_With_Append(t *testing.T) {
	// --- When ---
	buf := With([]byte{0, 1, 2}, Append)

	// --- Then ---
	assert.Exactly(t, os.O_APPEND, buf.flag)
	assert.Exactly(t, 3, buf.off)
	assert.Exactly(t, []byte{0, 1, 2}, buf.buf)
}

func Test_Buffer_tryGrowByReslice(t *testing.T) {
	tt := []struct {
		testN string

		len    int
		cap    int
		off    int
		grow   int
		expOK  bool
		expLen int
		expCap int
	}{
		{
			testN:  "1",
			len:    0,
			cap:    100,
			off:    0,
			grow:   50,
			expOK:  true,
			expLen: 50,
			expCap: 100,
		},
		{
			testN:  "2",
			len:    10,
			cap:    100,
			off:    10,
			grow:   50,
			expOK:  true,
			expLen: 60,
			expCap: 100,
		},
		{
			testN:  "3",
			len:    0,
			cap:    100,
			off:    0,
			grow:   100,
			expOK:  true,
			expLen: 100,
			expCap: 100,
		},
		{
			testN:  "4",
			len:    10,
			cap:    100,
			off:    10,
			grow:   90,
			expOK:  true,
			expLen: 100,
			expCap: 100,
		},
		{
			testN:  "5",
			len:    10,
			cap:    100,
			off:    10,
			grow:   150,
			expOK:  false,
			expLen: 10,
			expCap: 100,
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			data := make([]byte, tc.len, tc.cap)
			buf := With(data, Offset(tc.off))

			// --- When ---
			ok := buf.tryGrowByReslice(tc.grow)

			// --- Then ---
			assert.Exactly(t, tc.expOK, ok, "test %s", tc.testN)
			assert.Exactly(t, tc.off, buf.off, "test %s", tc.testN)
			assert.Exactly(t, tc.expLen, buf.Len(), "test %s", tc.testN)
			assert.Exactly(t, tc.expCap, buf.Cap(), "test %s", tc.testN)
		})
	}
}

func Test_Buffer_grow(t *testing.T) {
	tt := []struct {
		testN string

		len    int
		cap    int
		off    int
		grow   int
		expLen int
		expCap int
	}{
		{
			testN:  "1",
			len:    0,
			cap:    100,
			off:    0,
			grow:   50,
			expLen: 50,
			expCap: 100,
		},
		{
			testN:  "2",
			len:    10,
			cap:    100,
			off:    10,
			grow:   50,
			expLen: 60,
			expCap: 100,
		},
		{
			testN:  "3",
			len:    0,
			cap:    100,
			off:    0,
			grow:   100,
			expLen: 100,
			expCap: 100,
		},
		{
			testN:  "4",
			len:    10,
			cap:    100,
			off:    10,
			grow:   90,
			expLen: 100,
			expCap: 100,
		},
		{
			testN:  "5",
			len:    10,
			cap:    100,
			off:    5,
			grow:   150,
			expLen: 350,
			expCap: 350,
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			data := make([]byte, tc.len, tc.cap)
			buf := With(data, Offset(tc.off))

			// --- When ---
			buf.grow(tc.grow)

			// --- Then ---
			assert.Exactly(t, tc.off, buf.off, "test %s", tc.testN)
			assert.Exactly(t, tc.expLen, buf.Len(), "test %s", tc.testN)
			assert.Exactly(t, tc.expCap, buf.Cap(), "test %s", tc.testN)
		})
	}
}

func Test_Buffer_Grow(t *testing.T) {
	// --- Given ---
	data := make([]byte, 10, 15)
	buf := With(data, Offset(5))

	// --- When ---
	buf.Grow(20)

	// --- Then ---
	assert.Exactly(t, 10, buf.Len())
	assert.Exactly(t, 30, buf.Cap())
	assert.Exactly(t, 5, buf.Offset())
}

func Test_Buffer_Grow_AlreadyEnoughSpace(t *testing.T) {
	// --- Given ---
	data := make([]byte, 10, 15)
	buf := With(data, Offset(5))

	// --- When ---
	buf.Grow(5)

	// --- Then ---
	assert.Exactly(t, 10, buf.Len())
	assert.Exactly(t, 15, buf.Cap())
	assert.Exactly(t, 5, buf.Offset())
}

func Test_Buffer_Grow_Panics(t *testing.T) {
	// --- Given ---
	buf := &Buffer{}

	// --- Then ---
	assert.Panics(t, func() { buf.Grow(-1) })
}

func Test_Buffer_Write(t *testing.T) {
	tt := []struct {
		testN string

		init   []byte
		opts   []func(*Buffer)
		src    []byte
		expN   int
		expOff int
		expLen int
		expCap int
		expBuf []byte
	}{
		{
			testN:  "zero value",
			init:   nil,
			opts:   nil,
			src:    []byte{0, 1, 2},
			expN:   3,
			expOff: 3,
			expLen: 3,
			expCap: 64,
			expBuf: []byte{0, 1, 2},
		},
		{
			testN:  "empty with capacity",
			init:   make([]byte, 0, 5),
			opts:   nil,
			src:    []byte{0, 1, 2},
			expN:   3,
			expOff: 3,
			expLen: 3,
			expCap: 5,
			expBuf: []byte{0, 1, 2},
		},
		{
			testN:  "empty with capacity write more then cap",
			init:   make([]byte, 0, 5),
			opts:   nil,
			src:    []byte{0, 1, 2, 3, 4, 5},
			expN:   6,
			expOff: 6,
			expLen: 6,
			expCap: 16,
			expBuf: []byte{0, 1, 2, 3, 4, 5},
		},
		{
			testN:  "offset at len",
			init:   []byte{0, 1, 2},
			opts:   []func(*Buffer){Offset(3)},
			src:    []byte{3, 4, 5},
			expN:   3,
			expOff: 6,
			expLen: 6,
			expCap: 9,
			expBuf: []byte{0, 1, 2, 3, 4, 5},
		},
		{
			testN:  "append",
			init:   []byte{0, 1, 2},
			opts:   []func(*Buffer){Append},
			src:    []byte{3, 4, 5},
			expN:   3,
			expOff: 6,
			expLen: 6,
			expCap: 9,
			expBuf: []byte{0, 1, 2, 3, 4, 5},
		},
		{
			testN:  "override and extend",
			init:   []byte{0, 1, 2},
			opts:   []func(*Buffer){Offset(1)},
			src:    []byte{3, 4, 5},
			expN:   3,
			expOff: 4,
			expLen: 4,
			expCap: 9,
			expBuf: []byte{0, 3, 4, 5},
		},
		{
			testN:  "override and extend big",
			init:   []byte{0, 1, 2},
			opts:   []func(*Buffer){Offset(1)},
			src:    bytes.Repeat([]byte{0, 1}, 1<<20),
			expN:   2 * 1 << 20,
			expOff: 2*1<<20 + 1,
			expLen: 2*1<<20 + 1,
			expCap: 2*1<<20 + 6,
			expBuf: append([]byte{0}, bytes.Repeat([]byte{0, 1}, 1<<20)...),
		},
		{
			testN:  "override tail",
			init:   []byte{0, 1, 2},
			opts:   []func(*Buffer){Offset(1)},
			src:    []byte{3, 4},
			expN:   2,
			expOff: 3,
			expLen: 3,
			expCap: 3,
			expBuf: []byte{0, 3, 4},
		},
		{
			testN:  "override middle",
			init:   []byte{0, 1, 2, 3},
			opts:   []func(*Buffer){Offset(1)},
			src:    []byte{4, 5},
			expN:   2,
			expOff: 3,
			expLen: 4,
			expCap: 4,
			expBuf: []byte{0, 4, 5, 3},
		},
		{
			testN:  "override all no extend",
			init:   []byte{0, 1, 2, 3},
			opts:   nil,
			src:    []byte{4, 5, 6, 7},
			expN:   4,
			expOff: 4,
			expLen: 4,
			expCap: 4,
			expBuf: []byte{4, 5, 6, 7},
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			var buf *Buffer
			var err error

			if tc.init == nil {
				buf = &Buffer{} // Test for zero value.
			} else {
				buf = With(tc.init, tc.opts...)
			}

			// --- When ---
			n, err := buf.Write(tc.src)

			// --- Then ---
			assert.NoError(t, err, "test %s", tc.testN)
			assert.Exactly(t, tc.expN, n, "test %s", tc.testN)
			assert.Exactly(t, tc.expOff, buf.Offset(), "test %s", tc.testN)
			assert.Exactly(t, tc.expLen, buf.Len(), "test %s", tc.testN)
			assert.Exactly(t, tc.expCap, buf.Cap(), "test %s", tc.testN)
			assert.Exactly(t, tc.expBuf, buf.buf, "test %s", tc.testN)
			assert.NoError(t, buf.Close(), "test %s", tc.testN)
		})
	}
}

func Test_Buffer_WriteByte(t *testing.T) {
	tt := []struct {
		testN string

		init   []byte
		opts   []func(*Buffer)
		expOff int
		expLen int
		expCap int
		expBuf []byte
	}{
		{
			testN:  "zero value",
			init:   nil,
			opts:   nil,
			expOff: 1,
			expLen: 1,
			expCap: 64,
			expBuf: []byte{0xFF},
		},
		{
			testN:  "empty with capacity",
			init:   make([]byte, 0, 5),
			opts:   nil,
			expOff: 1,
			expLen: 1,
			expCap: 5,
			expBuf: []byte{0xFF},
		},
		{
			testN:  "offset at len",
			init:   []byte{0, 1, 2},
			opts:   []func(*Buffer){Offset(3)},
			expOff: 4,
			expLen: 4,
			expCap: 7,
			expBuf: []byte{0, 1, 2, 0xFF},
		},
		{
			testN:  "append",
			init:   []byte{0, 1, 2},
			opts:   []func(*Buffer){Append},
			expOff: 4,
			expLen: 4,
			expCap: 7,
			expBuf: []byte{0, 1, 2, 0xFF},
		},
		{
			testN:  "override tail",
			init:   []byte{0, 1, 2},
			opts:   []func(*Buffer){Offset(2)},
			expOff: 3,
			expLen: 3,
			expCap: 3,
			expBuf: []byte{0, 1, 0xFF},
		},
		{
			testN:  "override middle",
			init:   []byte{0, 1, 2, 3},
			opts:   []func(*Buffer){Offset(1)},
			expOff: 2,
			expLen: 4,
			expCap: 4,
			expBuf: []byte{0, 0xFF, 2, 3},
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			var buf *Buffer
			var err error

			if tc.init == nil {
				buf = &Buffer{} // Test for zero value.
			} else {
				buf = With(tc.init, tc.opts...)
			}

			// --- When ---
			err = buf.WriteByte(0xFF)

			// --- Then ---
			assert.NoError(t, err, "test %s", tc.testN)
			assert.Exactly(t, tc.expOff, buf.Offset(), "test %s", tc.testN)
			assert.Exactly(t, tc.expLen, buf.Len(), "test %s", tc.testN)
			assert.Exactly(t, tc.expCap, buf.Cap(), "test %s", tc.testN)
			assert.Exactly(t, tc.expBuf, buf.buf, "test %s", tc.testN)
			assert.NoError(t, buf.Close(), "test %s", tc.testN)
		})
	}
}

func Test_Buffer_WriteAt(t *testing.T) {
	tt := []struct {
		testN string

		init   []byte
		opts   []func(*Buffer)
		src    []byte
		off    int64
		expN   int
		expOff int
		expLen int
		expCap int
		expBuf []byte
	}{
		{
			testN:  "zero value - write at zero offset",
			init:   nil,
			opts:   nil,
			src:    []byte{0, 1, 2},
			off:    0,
			expN:   3,
			expOff: 0,
			expLen: 3,
			expCap: 64,
			expBuf: []byte{0, 1, 2},
		},
		{
			testN:  "write at zero offset - override",
			init:   []byte{0, 1, 2},
			opts:   nil,
			src:    []byte{3, 4, 5},
			off:    0,
			expN:   3,
			expOff: 0,
			expLen: 3,
			expCap: 3,
			expBuf: []byte{3, 4, 5},
		},
		{
			testN:  "write at offset middle - no extend",
			init:   []byte{0, 1, 2},
			opts:   nil,
			src:    []byte{3, 4},
			off:    1,
			expN:   2,
			expOff: 0,
			expLen: 3,
			expCap: 3,
			expBuf: []byte{0, 3, 4},
		},
		{
			testN:  "write at offset middle - extend",
			init:   []byte{0, 1, 2},
			opts:   nil,
			src:    []byte{3, 4, 5},
			off:    1,
			expN:   3,
			expOff: 0,
			expLen: 4,
			expCap: 7,
			expBuf: []byte{0, 3, 4, 5},
		},
		{
			testN:  "append",
			init:   []byte{0, 1, 2},
			opts:   nil,
			src:    []byte{3, 4, 5},
			off:    3,
			expN:   3,
			expOff: 0,
			expLen: 6,
			expCap: 9,
			expBuf: []byte{0, 1, 2, 3, 4, 5},
		},
		{
			testN:  "write at offset beyond len - within cap",
			init:   make([]byte, 3, 6),
			opts:   nil,
			src:    []byte{1, 2},
			off:    4,
			expN:   2,
			expOff: 0,
			expLen: 6,
			expCap: 6,
			expBuf: []byte{0, 0, 0, 0, 1, 2},
		},
		{
			testN:  "write at offset beyond len - beyond cap",
			init:   make([]byte, 3, 6),
			opts:   nil,
			src:    []byte{1, 2},
			off:    5,
			expN:   2,
			expOff: 0,
			expLen: 7,
			expCap: 16,
			expBuf: []byte{0, 0, 0, 0, 0, 1, 2},
		},
		{
			testN:  "write at offset beyond cap",
			init:   make([]byte, 3, 6),
			opts:   nil,
			src:    []byte{1, 2},
			off:    8,
			expN:   2,
			expOff: 0,
			expLen: 10,
			expCap: 19,
			expBuf: []byte{0, 0, 0, 0, 0, 0, 0, 0, 1, 2},
		},
		{
			testN:  "write at offset beyond cap - offset close to len",
			init:   make([]byte, 5, 7),
			opts:   []func(*Buffer){Offset(4)},
			src:    []byte{1, 2},
			off:    8,
			expN:   2,
			expOff: 4,
			expLen: 10,
			expCap: 19,
			expBuf: []byte{0, 0, 0, 0, 0, 0, 0, 0, 1, 2},
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			var buf *Buffer
			var err error

			if tc.init == nil {
				buf = &Buffer{} // Test for zero value.
			} else {
				buf = With(tc.init, tc.opts...)
			}

			// --- When ---
			n, err := buf.WriteAt(tc.src, tc.off)

			// --- Then ---
			assert.NoError(t, err, "test %s", tc.testN)
			assert.Exactly(t, tc.expN, n, "test %s", tc.testN)
			assert.Exactly(t, tc.expOff, buf.Offset(), "test %s", tc.testN)
			assert.Exactly(t, tc.expLen, buf.Len(), "test %s", tc.testN)
			assert.Exactly(t, tc.expCap, buf.Cap(), "test %s", tc.testN)
			assert.Exactly(t, tc.expBuf, buf.buf, "test %s", tc.testN)
			assert.NoError(t, buf.Close(), "test %s", tc.testN)
		})
	}
}

func Test_Buffer_ReadFrom(t *testing.T) {
	tt := []struct {
		testN string

		init   []byte
		opts   []func(*Buffer)
		src    []byte
		expN   int64
		expOff int
		expLen int
		expCap int
		expBuf []byte
	}{
		{
			testN:  "zero value",
			init:   nil,
			opts:   nil,
			src:    bytes.Repeat([]byte{1, 2, 3}, 1<<9),
			expN:   3 * 1 << 9,
			expOff: 3 * 1 << 9,
			expLen: 3 * 1 << 9,
			expCap: 3584,
			expBuf: bytes.Repeat([]byte{1, 2, 3}, 1<<9),
		},
		{
			testN:  "append",
			init:   []byte{0, 1, 2},
			opts:   []func(*Buffer){Append},
			src:    []byte{3, 4, 5},
			expN:   3,
			expOff: 6,
			expLen: 6,
			expCap: 518,
			expBuf: []byte{0, 1, 2, 3, 4, 5},
		},
		{
			testN:  "read up to len",
			init:   make([]byte, 3, 6),
			opts:   nil,
			src:    []byte{0, 1, 2},
			expN:   3,
			expOff: 3,
			expLen: 3,
			expCap: 524,
			expBuf: []byte{0, 1, 2},
		},
		{
			testN:  "read up to cap",
			init:   make([]byte, 3, 6),
			opts:   []func(*Buffer){Append},
			src:    []byte{3, 4, 5},
			expN:   3,
			expOff: 6,
			expLen: 6,
			expCap: 524,
			expBuf: []byte{0, 0, 0, 3, 4, 5},
		},
		{
			testN:  "use of tmp space",
			init:   bytes.Repeat([]byte{0}, 50),
			opts:   []func(*Buffer){Offset(25)},
			src:    bytes.Repeat([]byte{1, 2, 3}, 1<<9),
			expN:   3 * 1 << 9,
			expOff: 3*1<<9 + 25,
			expLen: 3*1<<9 + 25,
			expCap: 3984,
			expBuf: append(bytes.Repeat([]byte{0}, 25), bytes.Repeat([]byte{1, 2, 3}, 1<<9)...),
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			var buf *Buffer
			var err error

			if tc.init == nil {
				buf = &Buffer{} // Test for zero value.
			} else {
				buf = With(tc.init, tc.opts...)
			}

			// --- When ---
			n, err := buf.ReadFrom(bytes.NewReader(tc.src))

			// --- Then ---
			assert.NoError(t, err, "test %s", tc.testN)
			assert.Exactly(t, tc.expN, n, "test %s", tc.testN)
			assert.Exactly(t, tc.expOff, buf.Offset(), "test %s", tc.testN)
			assert.Exactly(t, tc.expLen, buf.Len(), "test %s", tc.testN)
			assert.Exactly(t, tc.expCap, buf.Cap(), "test %s", tc.testN)
			assert.Exactly(t, tc.expBuf, buf.buf, "test %s", tc.testN)
			assert.NoError(t, buf.Close(), "test %s", tc.testN)
		})
	}
}

// /////////////////////////////////////////////////////////////////////////////

func Test_Buffer_WriteAt_ZeroValue(t *testing.T) {
	// --- Given ---
	buf := &Buffer{}

	// --- When ---
	n, err := buf.WriteAt([]byte{0, 1, 2}, 0)

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, 3, n)
	assert.Exactly(t, 0, buf.Offset())
	assert.Exactly(t, 3, buf.Len())
	assert.Exactly(t, 64, buf.Cap())
	want := []byte{0, 1, 2}
	assert.Exactly(t, want, buf.buf)
	assert.NoError(t, buf.Close())
}

func Test_Buffer_WriteAt_OverrideAndExtend(t *testing.T) {
	// --- Given ---
	buf := With([]byte{0, 1, 2})

	// --- When ---
	data := bytes.Repeat([]byte{0, 1}, 500)
	n, err := buf.WriteAt(data, 1)

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, 1000, n)
	assert.Exactly(t, 0, buf.Offset())
	assert.Exactly(t, 1001, buf.Len())
	assert.Exactly(t, 1004, buf.Cap())
	want := append([]byte{0}, bytes.Repeat([]byte{0, 1}, 500)...)
	assert.Exactly(t, want, buf.buf)
	assert.NoError(t, buf.Close())
}

func Test_Buffer_WriteAt_BeyondCap(t *testing.T) {
	// --- Given ---
	buf := With([]byte{0, 1, 2})

	// --- When ---
	n, err := buf.WriteAt([]byte{3, 4, 5}, 1000)

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, 3, n)
	assert.Exactly(t, 0, buf.Offset())
	assert.Exactly(t, 1003, buf.Len())
	assert.Exactly(t, 1006, buf.Cap())
	want := append([]byte{0, 1, 2}, bytes.Repeat([]byte{0}, 997)...)
	want = append(want, []byte{3, 4, 5}...)
	assert.Exactly(t, want, buf.buf)
	assert.NoError(t, buf.Close())
}

func Test_Buffer_WriteTo(t *testing.T) {
	// --- Given ---
	buf := With([]byte{0, 1, 2, 3}, Offset(1))

	// --- When ---
	dst := &bytes.Buffer{}
	n, err := buf.WriteTo(dst)

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, int64(3), n)
	assert.Exactly(t, []byte{1, 2, 3}, dst.Bytes())
	assert.Exactly(t, 4, buf.Offset())
}

func Test_Buffer_WriteTo_OffsetAtTheEnd(t *testing.T) {
	// --- Given ---
	buf := With([]byte{0, 1, 2, 3}, Offset(4))

	// --- When ---
	dst := &bytes.Buffer{}
	n, err := buf.WriteTo(dst)

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, int64(0), n)
	assert.Exactly(t, []byte(nil), dst.Bytes())
	assert.Exactly(t, 4, buf.Offset())
}

func Test_Buffer_WriteString(t *testing.T) {
	// --- Given ---
	buf := With([]byte{0, 1, 2}, Offset(1))

	// --- When ---
	n, err := buf.WriteString("abc")

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, 3, n)
	assert.Exactly(t, []byte{0, 0x61, 0x62, 0x63}, buf.buf)
	assert.Exactly(t, 4, buf.Offset())
}

func Test_Buffer_Read_ZeroValue(t *testing.T) {
	// --- Given ---
	buf := &Buffer{}

	// --- When ---
	dst := make([]byte, 3)
	n, err := buf.Read(dst)

	// --- Then ---
	assert.ErrorIs(t, err, io.EOF)
	assert.Exactly(t, 0, n)
	assert.Exactly(t, 0, buf.Offset())
	assert.Exactly(t, 0, buf.Len())
	assert.Exactly(t, 0, buf.Cap())
	want := []byte{0, 0, 0}
	assert.Exactly(t, want, dst)
	assert.NoError(t, buf.Close())
}

func Test_Buffer_Read_WithSmallBuffer(t *testing.T) {
	// --- Given ---
	buf := With([]byte{0, 1, 2, 3, 4})
	dst := make([]byte, 3)

	// --- Then ---

	// First read.
	n, err := buf.Read(dst)

	assert.NoError(t, err)
	assert.Exactly(t, 3, n)
	assert.Exactly(t, 3, buf.Offset())
	assert.Exactly(t, 5, buf.Len())
	assert.Exactly(t, 5, buf.Cap())
	want := []byte{0, 1, 2}
	assert.Exactly(t, want, dst)

	// Second read.
	n, err = buf.Read(dst)

	assert.NoError(t, err)
	assert.Exactly(t, 2, n)
	assert.Exactly(t, 5, buf.Offset())
	assert.Exactly(t, 5, buf.Len())
	assert.Exactly(t, 5, buf.Cap())
	want = []byte{3, 4, 2}
	assert.Exactly(t, want, dst)

	// Third read.
	n, err = buf.Read(dst)

	assert.ErrorIs(t, err, io.EOF)
	assert.Exactly(t, 0, n)
	assert.Exactly(t, 5, buf.Offset())
	assert.Exactly(t, 5, buf.Len())
	assert.Exactly(t, 5, buf.Cap())
	want = []byte{3, 4, 2}
	assert.Exactly(t, want, dst)

	assert.NoError(t, buf.Close())
}

func Test_Buffer_Read_BeyondLen(t *testing.T) {
	// --- Given ---
	buf := With([]byte{0, 1, 2})
	_, err := buf.Seek(5, io.SeekStart)
	require.NoError(t, err)

	// --- When ---
	dst := make([]byte, 3)
	n, err := buf.Read(dst)

	// --- Then ---
	assert.ErrorIs(t, err, io.EOF)
	assert.Exactly(t, 0, n)
	assert.Exactly(t, 5, buf.Offset())
	assert.Exactly(t, 3, buf.Len())
	assert.Exactly(t, 3, buf.Cap())
	want := []byte{0, 0, 0}
	assert.Exactly(t, want, dst)
	assert.NoError(t, buf.Close())
}

func Test_Buffer_Read_BigBuffer(t *testing.T) {
	// --- Given ---
	buf := With([]byte{0, 1, 2})

	// --- When ---
	dst := make([]byte, 6)
	n, err := buf.Read(dst)

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, 3, n)
	assert.Exactly(t, 3, buf.Offset())
	assert.Exactly(t, 3, buf.Len())
	assert.Exactly(t, 3, buf.Cap())
	want := []byte{0, 1, 2, 0, 0, 0}
	assert.Exactly(t, want, dst)
	assert.NoError(t, buf.Close())
}

func Test_Buffer_Read(t *testing.T) {
	tt := []struct {
		testN string

		init   []byte
		opts   []func(*Buffer)
		dst    []byte
		expN   int
		expOff int
		expLen int
		expCap int
		expDst []byte
	}{
		{
			testN:  "read all",
			init:   []byte{0, 1, 2},
			opts:   nil,
			dst:    make([]byte, 3, 3),
			expN:   3,
			expOff: 3,
			expLen: 3,
			expCap: 3,
			expDst: []byte{0, 1, 2},
		},
		{
			testN:  "read head",
			init:   []byte{0, 1, 2},
			opts:   nil,
			dst:    make([]byte, 2, 3),
			expN:   2,
			expOff: 2,
			expLen: 3,
			expCap: 3,
			expDst: []byte{0, 1},
		},
		{
			testN:  "read tail",
			init:   []byte{0, 1, 2},
			opts:   []func(*Buffer){Offset(1)},
			dst:    make([]byte, 2, 3),
			expN:   2,
			expOff: 3,
			expLen: 3,
			expCap: 3,
			expDst: []byte{1, 2},
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			buf := With(tc.init, tc.opts...)

			// --- When ---
			n, err := buf.Read(tc.dst)

			// --- Then ---
			assert.NoError(t, err, "test %s", tc.testN)
			assert.Exactly(t, tc.expN, n, "test %s", tc.testN)
			assert.Exactly(t, tc.expOff, buf.Offset(), "test %s", tc.testN)
			assert.Exactly(t, tc.expLen, buf.Len(), "test %s", tc.testN)
			assert.Exactly(t, tc.expCap, buf.Cap(), "test %s", tc.testN)
			assert.Exactly(t, tc.expDst, tc.dst, "test %s", tc.testN)
			assert.NoError(t, buf.Close(), "test %s", tc.testN)
		})
	}
}

func Test_Buffer_ReadByte(t *testing.T) {
	// --- Given ---
	buf := With([]byte{0, 1, 2}, Offset(2))

	// --- When ---
	got, err := buf.ReadByte()

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, 3, buf.Offset())
	assert.Exactly(t, 3, buf.Len())
	assert.Exactly(t, 3, buf.Cap())
	assert.Exactly(t, byte(2), got)
	assert.NoError(t, buf.Close())
}

func Test_Buffer_ReadByte_EOF(t *testing.T) {
	// --- Given ---
	buf := With([]byte{0, 1, 2}, Offset(3))

	// --- When ---
	got, err := buf.ReadByte()

	// --- Then ---
	assert.ErrorIs(t, err, io.EOF)
	assert.Exactly(t, 3, buf.Offset())
	assert.Exactly(t, 3, buf.Len())
	assert.Exactly(t, 3, buf.Cap())
	assert.Exactly(t, byte(0), got)
	assert.NoError(t, buf.Close())
}

func Test_Buffer_ReadAt_BeyondLen(t *testing.T) {
	// --- Given ---
	buf := With([]byte{0, 1, 2})

	// --- When ---
	dst := make([]byte, 4)
	n, err := buf.ReadAt(dst, 6)

	// --- Then ---
	assert.ErrorIs(t, err, io.EOF)
	assert.Exactly(t, 0, n)
	assert.Exactly(t, 0, buf.Offset())
	assert.Exactly(t, 3, buf.Len())
	assert.Exactly(t, 3, buf.Cap())
	want := []byte{0, 0, 0, 0}
	assert.Exactly(t, want, dst)
	assert.NoError(t, buf.Close())
}

func Test_Buffer_ReadAt_BigBuffer(t *testing.T) {
	// --- Given ---
	buf := With([]byte{0, 1, 2}, Offset(1))
	dst := make([]byte, 4)

	// --- When ---
	n, err := buf.ReadAt(dst, 0)

	// --- Then ---
	assert.ErrorIs(t, err, io.EOF)
	assert.Exactly(t, 3, n)
	assert.Exactly(t, 1, buf.Offset())
	assert.Exactly(t, 3, buf.Len())
	assert.Exactly(t, 3, buf.Cap())
	want := []byte{0, 1, 2, 0}
	assert.Exactly(t, want, dst)
	assert.NoError(t, buf.Close())
}

func Test_Buffer_ReadAt(t *testing.T) {
	tt := []struct {
		testN string

		init   []byte
		opts   []func(*Buffer)
		dst    []byte
		off    int64
		expN   int
		expOff int
		expLen int
		expCap int
		expDst []byte
	}{
		{
			testN:  "read all",
			init:   []byte{0, 1, 2},
			opts:   []func(*Buffer){Offset(1)},
			dst:    make([]byte, 3),
			off:    0,
			expN:   3,
			expOff: 1,
			expLen: 3,
			expCap: 3,
			expDst: []byte{0, 1, 2},
		},
		{
			testN:  "read head",
			init:   []byte{0, 1, 2},
			opts:   []func(*Buffer){Offset(1)},
			dst:    make([]byte, 2, 3),
			off:    0,
			expN:   2,
			expOff: 1,
			expLen: 3,
			expCap: 3,
			expDst: []byte{0, 1},
		},
		{
			testN:  "read tail",
			init:   []byte{0, 1, 2},
			opts:   []func(*Buffer){Offset(2)},
			dst:    make([]byte, 2, 3),
			off:    1,
			expN:   2,
			expOff: 2,
			expLen: 3,
			expCap: 3,
			expDst: []byte{1, 2},
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			buf := With(tc.init, tc.opts...)

			// --- When ---
			n, err := buf.ReadAt(tc.dst, tc.off)

			// --- Then ---
			assert.NoError(t, err, "test %s", tc.testN)
			assert.Exactly(t, tc.expN, n, "test %s", tc.testN)
			assert.Exactly(t, tc.expOff, buf.Offset(), "test %s", tc.testN)
			assert.Exactly(t, tc.expLen, buf.Len(), "test %s", tc.testN)
			assert.Exactly(t, tc.expCap, buf.Cap(), "test %s", tc.testN)
			assert.Exactly(t, tc.expDst, tc.dst, "test %s", tc.testN)
			assert.NoError(t, buf.Close(), "test %s", tc.testN)
		})
	}
}

func Test_Buffer_String(t *testing.T) {
	// --- Given ---
	buf := With([]byte{'A', 'B', 'C', 'D'}, Offset(1))

	// --- When ---
	s := buf.String()

	// --- Then ---
	assert.Exactly(t, "BCD", s)
	assert.Exactly(t, 4, buf.Offset())
}

func Test_Buffer_String_ZeroValueBuffer(t *testing.T) {
	// --- Given ---
	buf := &Buffer{}

	// --- When ---
	s := buf.String()

	// --- Then ---
	assert.Exactly(t, "", s)
	assert.Exactly(t, 0, buf.Offset())
}

func Test_Buffer_Seek(t *testing.T) {
	// --- Given ---
	tt := []struct {
		testN string

		seek   int64
		whence int
		wantN  int64
		wantD  []byte
	}{
		{"1", 0, io.SeekCurrent, 1, []byte{1, 2, 3}},
		{"2", 0, io.SeekEnd, 4, []byte{}},
		{"3", -1, io.SeekEnd, 3, []byte{3}},
		{"4", -3, io.SeekEnd, 1, []byte{1, 2, 3}},
		{"5", 0, io.SeekStart, 0, []byte{0, 1, 2, 3}},
		{"6", 2, io.SeekStart, 2, []byte{2, 3}},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			buf := With([]byte{0, 1, 2, 3}, Offset(1))

			// --- When ---
			n, err := buf.Seek(tc.seek, tc.whence)

			// --- Then ---
			assert.NoError(t, err, "test %s", tc.testN)
			assert.Exactly(t, tc.wantN, n, "test %s", tc.testN)

			got, err := ioutil.ReadAll(buf)
			assert.NoError(t, err, "test %s", tc.testN)
			assert.Exactly(t, tc.wantD, got, "test %s", tc.testN)

			assert.NoError(t, buf.Close(), "test %s", tc.testN)
		})
	}
}

func Test_Buffer_Seek_NegativeFinalOffset(t *testing.T) {
	// --- Given ---
	buf := With([]byte{0, 1, 2})

	// --- When ---
	n, err := buf.Seek(-4, io.SeekEnd)

	// --- Then ---
	assert.ErrorIs(t, err, os.ErrInvalid)
	assert.Exactly(t, int64(0), n)
}

func Test_Buffer_Seek_BeyondLen(t *testing.T) {
	// --- Given ---
	buf := With([]byte{0, 1, 2})

	// --- When ---
	n, err := buf.Seek(5, io.SeekStart)

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, int64(5), n)
}

func Test_Buffer_SeekStart(t *testing.T) {
	// --- Given ---
	buf := With([]byte{0, 1, 2}, Offset(2))

	// --- When ---
	n := buf.SeekStart()

	// --- Then ---
	assert.Exactly(t, int64(2), n)
	assert.Exactly(t, 0, buf.off)
}

func Test_Buffer_Truncate(t *testing.T) {
	tt := []struct {
		testN string

		init   []byte
		opts   []func(*Buffer)
		off    int64
		expOff int
		expLen int
		expCap int
		expBuf []byte
	}{
		{
			testN:  "truncate to zero",
			init:   []byte{0, 1, 2, 3},
			opts:   nil,
			off:    0,
			expOff: 0,
			expLen: 0,
			expCap: 4,
			expBuf: []byte{},
		},
		{
			testN:  "truncate to one",
			init:   []byte{0, 1, 2, 3},
			opts:   nil,
			off:    1,
			expOff: 0,
			expLen: 1,
			expCap: 4,
			expBuf: []byte{0},
		},
		{
			testN:  "truncate beyond len, less then cap",
			init:   make([]byte, 3, 5),
			opts:   nil,
			off:    4,
			expOff: 0,
			expLen: 4,
			expCap: 5,
			expBuf: []byte{0, 0, 0, 0},
		},
		{
			testN:  "truncate beyond cap",
			init:   make([]byte, 3, 5),
			opts:   nil,
			off:    6,
			expOff: 0,
			expLen: 6,
			expCap: 13,
			expBuf: []byte{0, 0, 0, 0, 0, 0},
		},
		{
			testN:  "truncate at len",
			init:   make([]byte, 3, 5),
			opts:   nil,
			off:    3,
			expOff: 0,
			expLen: 3,
			expCap: 5,
			expBuf: []byte{0, 0, 0},
		},
		{
			testN:  "truncate at cap",
			init:   make([]byte, 3, 5),
			opts:   nil,
			off:    5,
			expOff: 0,
			expLen: 5,
			expCap: 5,
			expBuf: []byte{0, 0, 0, 0, 0},
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			var buf *Buffer
			var err error

			if tc.init == nil {
				buf = &Buffer{} // Test for zero value.
			} else {
				buf = With(tc.init, tc.opts...)
			}

			// --- When ---
			err = buf.Truncate(tc.off)

			// --- Then ---
			assert.NoError(t, err, "test %s", tc.testN)
			assert.Exactly(t, tc.expOff, buf.Offset(), "test %s", tc.testN)
			assert.Exactly(t, tc.expLen, buf.Len(), "test %s", tc.testN)
			assert.Exactly(t, tc.expCap, buf.Cap(), "test %s", tc.testN)
			assert.Exactly(t, tc.expBuf, buf.buf, "test %s", tc.testN)
			assert.NoError(t, buf.Close(), "test %s", tc.testN)
		})
	}
}

func Test_Buffer_Truncate_ToZeroAndWrite(t *testing.T) {
	// --- Given ---
	buf := With([]byte{0, 1, 2, 3})

	// --- When ---
	err := buf.Truncate(0)
	assert.NoError(t, err)

	n, err := buf.Write([]byte{4, 5})

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, 2, n)
	assert.Exactly(t, 2, buf.Offset())
	assert.Exactly(t, 2, buf.Len())
	assert.Exactly(t, 4, buf.Cap())
	assert.Exactly(t, []byte{4, 5}, buf.buf)
	assert.NoError(t, buf.Close())
}

func Test_Buffer_Truncate_BeyondLenAndWrite(t *testing.T) {
	// --- Given ---
	buf := With([]byte{0, 1, 2, 3}, Append)
	_, err := buf.Seek(1, io.SeekStart)
	require.NoError(t, err)

	// --- When ---
	assert.NoError(t, buf.Truncate(8))
	n, err := buf.Write([]byte{4, 5})

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, 2, n)
	assert.Exactly(t, 10, buf.Offset())
	assert.Exactly(t, 10, buf.Len())
	assert.Exactly(t, 12, buf.Cap())
	assert.Exactly(t, []byte{0, 1, 2, 3, 0, 0, 0, 0, 4, 5}, buf.buf)
	assert.NoError(t, buf.Close())
}

func Test_Buffer_Truncate_BeyondCapAndWrite(t *testing.T) {
	// --- Given ---
	data := make([]byte, 4, 8)
	data[0] = 0
	data[1] = 1
	data[2] = 2
	data[3] = 3
	buf := With(data, Append)

	// --- When ---
	assert.NoError(t, buf.Truncate(10))
	n, err := buf.Write([]byte{4, 5})

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, 2, n)
	assert.Exactly(t, 12, buf.Offset())
	assert.Exactly(t, 12, buf.Len())
	assert.Exactly(t, 22, buf.Cap())
	want := []byte{0, 1, 2, 3, 0, 0, 0, 0, 0, 0, 4, 5}
	assert.Exactly(t, want, buf.buf)
	assert.NoError(t, buf.Close())
}

func Test_Buffer_Truncate_ExtendBeyondLenResetAndWrite(t *testing.T) {
	// --- Given ---
	buf := With([]byte{0, 1, 2, 3}, Append)

	// --- When ---
	assert.NoError(t, buf.Truncate(8))
	assert.NoError(t, buf.Truncate(0))
	n, err := buf.Write([]byte{4, 5})

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, 2, n)
	assert.Exactly(t, 2, buf.Offset())
	assert.Exactly(t, 2, buf.Len())
	assert.Exactly(t, 12, buf.Cap())
	assert.Exactly(t, []byte{4, 5}, buf.buf)
	assert.NoError(t, buf.Close())
}

func Test_Buffer_Truncate_EdgeCaseWhenSizeEqualsLength(t *testing.T) {
	// --- Given ---
	buf := With([]byte{0, 1, 2, 3}, Append)

	// --- When ---
	assert.NoError(t, buf.Truncate(4))
	n, err := buf.Write([]byte{4, 5})

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, 2, n)
	assert.Exactly(t, 6, buf.Offset())
	assert.Exactly(t, 6, buf.Len())
	assert.Exactly(t, 10, buf.Cap())
	assert.Exactly(t, []byte{0, 1, 2, 3, 4, 5}, buf.buf)
	assert.NoError(t, buf.Close())
}

func Test_Buffer_Truncate_Error(t *testing.T) {
	// --- Given ---
	buf := With([]byte{0, 1, 2}, Append)

	// --- When ---
	err := buf.Truncate(-1)

	// --- Then ---
	assert.ErrorIs(t, err, os.ErrInvalid)
}

func Test_Buffer_Close_ZeroValue(t *testing.T) {
	// --- When ---
	buf := &Buffer{}

	// --- Then ---
	assert.NoError(t, buf.Close())
}

func Test_Buffer_Close_NilBuffer(t *testing.T) {
	// --- When ---
	var buf *Buffer

	// --- Then ---
	//goland:noinspection GoNilness
	assert.NoError(t, buf.Close())
}

func Test_Buffer_Release(t *testing.T) {
	// --- Given ---
	buf := With([]byte{0, 1, 2, 3}, Offset(1))

	// --- When ---
	got := buf.Release()

	// --- Then ---
	assert.Exactly(t, []byte{0, 1, 2, 3}, got)
	assert.Exactly(t, 0, buf.off)
	assert.Nil(t, buf.buf)
}

func Test_helpers_zeroOutSlice(t *testing.T) {
	// --- Given ---
	data := []byte{0, 1, 2, 3}

	// --- When ---
	zeroOutSlice(data)

	// --- Then ---
	assert.Exactly(t, []byte{0, 0, 0, 0}, data)
}
