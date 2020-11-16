package flexbuf

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_With(t *testing.T) {
	// --- When ---
	buf, err := With([]byte{0, 1, 2})

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, 0, buf.off)
	assert.Exactly(t, []byte{0, 1, 2}, buf.buf)
}

func Test_With_Offset(t *testing.T) {
	// --- When ---
	buf, err := With([]byte{0, 1, 2}, Offset(1))

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, 1, buf.off)
	assert.Exactly(t, []byte{0, 1, 2}, buf.buf)
}

func Test_With_Append(t *testing.T) {
	// --- When ---
	buf, err := With([]byte{0, 1, 2}, Append)

	// --- Then ---
	assert.NoError(t, err)
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
			buf, err := With(data, Offset(tc.off))
			require.NoError(t, err, "test %s", tc.testN)

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
			expLen: 155,
			expCap: 155,
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			data := make([]byte, tc.len, tc.cap)
			buf, err := With(data, Offset(tc.off))
			require.NoError(t, err, "test %s", tc.testN)

			// --- When ---
			buf.grow(tc.grow)

			// --- Then ---
			assert.Exactly(t, tc.off, buf.off, "test %s", tc.testN)
			assert.Exactly(t, tc.expLen, buf.Len(), "test %s", tc.testN)
			assert.Exactly(t, tc.expCap, buf.Cap(), "test %s", tc.testN)
		})
	}
}

func Test_Buffer_ReadFrom_ToEmpty(t *testing.T) {
	// --- Given ---
	src := bytes.NewBuffer(bytes.Repeat([]byte{1, 2}, 500))
	buf, err := New()
	require.NoError(t, err)

	// --- When ---
	n, err := buf.ReadFrom(src)

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, int64(1000), n)
	assert.Exactly(t, 1000, buf.Offset())
	assert.Exactly(t, 1000, buf.Len())
	assert.Exactly(t, 1512, buf.Cap())
	assert.Exactly(t, bytes.Repeat([]byte{1, 2}, 500), buf.buf)
	assert.NoError(t, buf.Close())
}

func Test_Buffer_ReadFrom_ToFull(t *testing.T) {
	// --- Given ---
	src := bytes.NewBuffer([]byte{3, 4, 5})
	buf, err := With([]byte{0, 1, 2}, Append)
	require.NoError(t, err)

	// --- When ---
	n, err := buf.ReadFrom(src)

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, int64(3), n)
	assert.Exactly(t, 6, buf.Offset())
	assert.Exactly(t, 6, buf.Len())
	assert.Exactly(t, 518, buf.Cap())
	want := []byte{0, 1, 2, 3, 4, 5}
	assert.Exactly(t, want, buf.buf)
	assert.NoError(t, buf.Close())
}

func Test_Buffer_Write(t *testing.T) {
	// --- Given ---
	buf := &Buffer{}

	// --- When ---
	n, err := buf.Write([]byte{0, 1, 2})

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, 3, n)
	assert.Exactly(t, 3, buf.Offset())
	assert.Exactly(t, 3, buf.Len())
	assert.Exactly(t, 3, buf.Cap())
	want := []byte{0, 1, 2}
	assert.Exactly(t, want, buf.buf)
	assert.NoError(t, buf.Close())
}

func Test_Buffer_Write_OverrideAndExtend(t *testing.T) {
	// --- Given ---
	data := bytes.Repeat([]byte{0, 1}, 500)
	buf, err := With([]byte{0, 1, 2}, Offset(1))
	require.NoError(t, err)

	// --- When ---
	n, err := buf.Write(data)

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, 1000, n)
	assert.Exactly(t, 1001, buf.Offset())
	assert.Exactly(t, 1001, buf.Len())
	assert.Exactly(t, 1001, buf.Cap())
	want := append([]byte{0}, data...)
	assert.Exactly(t, want, buf.buf)
	assert.NoError(t, buf.Close())
}

func Test_Buffer_Write_(t *testing.T) {
	tt := []struct {
		testN string

		init   []byte
		opts   []func(*Buffer) error
		src    []byte
		expN   int
		expOff int
		expLen int
		expCap int
		expBuf []byte
	}{
		{
			testN:  "append",
			init:   []byte{0, 1, 2},
			opts:   []func(*Buffer) error{Append},
			src:    []byte{3, 4, 5},
			expN:   3,
			expOff: 6,
			expLen: 6,
			expCap: 6,
			expBuf: []byte{0, 1, 2, 3, 4, 5},
		},
		{
			testN:  "override and extend",
			init:   []byte{0, 1, 2},
			opts:   []func(*Buffer) error{Offset(1)},
			src:    []byte{3, 4, 5},
			expN:   3,
			expOff: 4,
			expLen: 4,
			expCap: 4,
			expBuf: []byte{0, 3, 4, 5},
		},
		{
			testN:  "override tail",
			init:   []byte{0, 1, 2},
			opts:   []func(*Buffer) error{Offset(1)},
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
			opts:   []func(*Buffer) error{Offset(1)},
			src:    []byte{4, 5},
			expN:   2,
			expOff: 3,
			expLen: 4,
			expCap: 4,
			expBuf: []byte{0, 4, 5, 3},
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			buf, err := With(tc.init, tc.opts...)
			require.NoError(t, err)

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
