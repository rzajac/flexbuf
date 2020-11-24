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
	// --- When ---
	buf, err := New(Offset(-1))

	// --- Then ---
	assert.ErrorIs(t, err, ErrOutOfBounds)
	assert.Nil(t, buf)
}

func Test_With(t *testing.T) {
	// --- When ---
	buf, err := With([]byte{0, 1, 2})

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, 0, buf.flag)
	assert.Exactly(t, 0, buf.off)
	assert.Exactly(t, []byte{0, 1, 2}, buf.buf)
}

func Test_With_Offset(t *testing.T) {
	// --- When ---
	buf, err := With([]byte{0, 1, 2}, Offset(1))

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, 0, buf.flag)
	assert.Exactly(t, 1, buf.off)
	assert.Exactly(t, []byte{0, 1, 2}, buf.buf)
}

func Test_With_Offset_Negative(t *testing.T) {
	// --- When ---
	buf, err := With([]byte{0, 1, 2}, Offset(-1))

	// --- Then ---
	assert.ErrorIs(t, err, ErrOutOfBounds)
	assert.Nil(t, buf)
}

func Test_With_Offset_BeyondLen(t *testing.T) {
	// --- When ---
	buf, err := With([]byte{0, 1, 2}, Offset(5))

	// --- Then ---
	assert.ErrorIs(t, err, ErrOutOfBounds)
	assert.Nil(t, buf)
}

func Test_With_Append(t *testing.T) {
	// --- When ---
	buf, err := With([]byte{0, 1, 2}, Append)

	// --- Then ---
	assert.NoError(t, err)
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

func Test_Buffer_Write_ZeroValue(t *testing.T) {
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

func Test_Buffer_Write(t *testing.T) {
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

func Test_Buffer_WriteByte(t *testing.T) {
	// --- Given ---
	buf, err := With([]byte{0, 1, 2}, Offset(1))
	require.NoError(t, err)

	// --- When ---
	err = buf.WriteByte(3)

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, 2, buf.Offset())
	assert.Exactly(t, 3, buf.Len())
	assert.Exactly(t, 3, buf.Cap())
	want := []byte{0, 3, 2}
	assert.Exactly(t, want, buf.buf)
	assert.NoError(t, buf.Close())
}

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
	assert.Exactly(t, 3, buf.Cap())
	want := []byte{0, 1, 2}
	assert.Exactly(t, want, buf.buf)
	assert.NoError(t, buf.Close())
}

func Test_Buffer_WriteAt_OverrideAndExtend(t *testing.T) {
	// --- Given ---
	buf, err := With([]byte{0, 1, 2})
	require.NoError(t, err)

	// --- When ---
	data := bytes.Repeat([]byte{0, 1}, 500)
	n, err := buf.WriteAt(data, 1)

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, 1000, n)
	assert.Exactly(t, 0, buf.Offset())
	assert.Exactly(t, 1001, buf.Len())
	assert.Exactly(t, 1001, buf.Cap())
	want := append([]byte{0}, bytes.Repeat([]byte{0, 1}, 500)...)
	assert.Exactly(t, want, buf.buf)
	assert.NoError(t, buf.Close())
}

func Test_Buffer_WriteAt_BeyondCap(t *testing.T) {
	// --- Given ---
	buf, err := With([]byte{0, 1, 2})
	require.NoError(t, err)

	// --- When ---
	n, err := buf.WriteAt([]byte{3, 4, 5}, 1000)

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, 3, n)
	assert.Exactly(t, 0, buf.Offset())
	assert.Exactly(t, 1003, buf.Len())
	assert.Exactly(t, 1003, buf.Cap())
	want := append([]byte{0, 1, 2}, bytes.Repeat([]byte{0}, 997)...)
	want = append(want, []byte{3, 4, 5}...)
	assert.Exactly(t, want, buf.buf)
	assert.NoError(t, buf.Close())
}

func Test_Buffer_WriteAt(t *testing.T) {
	tt := []struct {
		testN string

		init   []byte
		opts   []func(*Buffer) error
		src    []byte
		off    int64
		expN   int
		expOff int
		expLen int
		expCap int
		expBuf []byte
	}{
		{
			testN:  "append",
			init:   []byte{0, 1, 2},
			opts:   nil,
			src:    []byte{3, 4, 5},
			off:    3,
			expN:   3,
			expOff: 0,
			expLen: 6,
			expCap: 6,
			expBuf: []byte{0, 1, 2, 3, 4, 5},
		},
		{
			testN:  "override and extend",
			init:   []byte{0, 1, 2},
			opts:   []func(*Buffer) error{Offset(2)},
			src:    []byte{3, 4, 5},
			off:    1,
			expN:   3,
			expOff: 2,
			expLen: 4,
			expCap: 4,
			expBuf: []byte{0, 3, 4, 5},
		},
		{
			testN:  "override tail",
			init:   []byte{0, 1, 2},
			opts:   []func(*Buffer) error{Offset(2)},
			src:    []byte{3, 4},
			off:    1,
			expN:   2,
			expOff: 2,
			expLen: 3,
			expCap: 3,
			expBuf: []byte{0, 3, 4},
		},
		{
			testN:  "override middle",
			init:   []byte{0, 1, 2, 3},
			opts:   []func(*Buffer) error{Offset(2)},
			src:    []byte{4, 5},
			off:    1,
			expN:   2,
			expOff: 2,
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

func Test_Buffer_WriteTo(t *testing.T) {
	// --- Given ---
	buf, err := With([]byte{0, 1, 2, 3}, Offset(1))
	require.NoError(t, err)

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
	buf, err := With([]byte{0, 1, 2, 3}, Offset(4))
	require.NoError(t, err)

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
	buf, err := With([]byte{0, 1, 2}, Offset(1))
	require.NoError(t, err)

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
	buf, err := With([]byte{0, 1, 2, 3, 4})
	require.NoError(t, err)
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
	buf, err := With([]byte{0, 1, 2})
	require.NoError(t, err)
	_, err = buf.Seek(5, io.SeekStart)
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
	buf, err := With([]byte{0, 1, 2})
	require.NoError(t, err)

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
		opts   []func(*Buffer) error
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
			opts:   []func(*Buffer) error{Offset(1)},
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
			buf, err := With(tc.init, tc.opts...)
			require.NoError(t, err)

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
	buf, err := With([]byte{0, 1, 2}, Offset(2))
	require.NoError(t, err)

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
	buf, err := With([]byte{0, 1, 2}, Offset(3))
	require.NoError(t, err)

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
	buf, err := With([]byte{0, 1, 2})
	require.NoError(t, err)

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
	buf, err := With([]byte{0, 1, 2}, Offset(1))
	require.NoError(t, err)
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
		opts   []func(*Buffer) error
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
			opts:   []func(*Buffer) error{Offset(1)},
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
			opts:   []func(*Buffer) error{Offset(1)},
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
			opts:   []func(*Buffer) error{Offset(2)},
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
			buf, err := With(tc.init, tc.opts...)
			require.NoError(t, err)

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
	buf, err := With([]byte{'A', 'B', 'C', 'D'}, Offset(1))
	require.NoError(t, err)

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
			buf, err := With([]byte{0, 1, 2, 3}, Offset(1))
			require.NoError(t, err, "test %s", tc.testN)

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
	buf, err := With([]byte{0, 1, 2})
	require.NoError(t, err)

	// --- When ---
	n, err := buf.Seek(-4, io.SeekEnd)

	// --- Then ---
	assert.ErrorIs(t, err, os.ErrInvalid)
	assert.Exactly(t, int64(0), n)
}

func Test_Buffer_Seek_BeyondLen(t *testing.T) {
	// --- Given ---
	buf, err := With([]byte{0, 1, 2})
	require.NoError(t, err)

	// --- When ---
	n, err := buf.Seek(5, io.SeekStart)

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, int64(5), n)
}

func Test_Buffer_Truncate_ToZero(t *testing.T) {
	// --- Given ---
	buf, err := With([]byte{0, 1, 2, 3})
	require.NoError(t, err)

	// --- When ---
	err = buf.Truncate(0)

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, 0, buf.Offset())
	assert.Exactly(t, 0, buf.Len())
	assert.Exactly(t, 4, buf.Cap())
	assert.Exactly(t, []byte{}, buf.buf)
	assert.NoError(t, buf.Close())
}

func Test_Buffer_Truncate_ToOne(t *testing.T) {
	// --- Given ---
	buf, err := With([]byte{0, 1, 2, 3})
	require.NoError(t, err)

	// --- When ---
	err = buf.Truncate(1)

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, 0, buf.Offset())
	assert.Exactly(t, 1, buf.Len())
	assert.Exactly(t, 4, buf.Cap())
	assert.Exactly(t, []byte{0}, buf.buf)
	assert.NoError(t, buf.Close())
}

func Test_Buffer_Truncate_ToZeroAndWrite(t *testing.T) {
	// --- Given ---
	buf, err := With([]byte{0, 1, 2, 3})
	require.NoError(t, err)

	// --- When ---
	err = buf.Truncate(0)
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
	buf, err := With([]byte{0, 1, 2, 3}, Append)
	require.NoError(t, err)
	_, err = buf.Seek(1, io.SeekStart)
	require.NoError(t, err)

	// --- When ---
	assert.NoError(t, buf.Truncate(8))
	n, err := buf.Write([]byte{4, 5})

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, 2, n)
	assert.Exactly(t, 10, buf.Offset())
	assert.Exactly(t, 10, buf.Len())
	assert.Exactly(t, 10, buf.Cap())
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
	buf, err := With(data, Append)
	require.NoError(t, err)

	// --- When ---
	assert.NoError(t, buf.Truncate(10))
	n, err := buf.Write([]byte{4, 5})

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, 2, n)
	assert.Exactly(t, 12, buf.Offset())
	assert.Exactly(t, 12, buf.Len())
	assert.Exactly(t, 12, buf.Cap())
	want := []byte{0, 1, 2, 3, 0, 0, 0, 0, 0, 0, 4, 5}
	assert.Exactly(t, want, buf.buf)
	assert.NoError(t, buf.Close())
}

func Test_Buffer_Truncate_ExtendBeyondLenResetAndWrite(t *testing.T) {
	// --- Given ---
	buf, err := With([]byte{0, 1, 2, 3}, Append)
	require.NoError(t, err)

	// --- When ---
	assert.NoError(t, buf.Truncate(8))
	assert.NoError(t, buf.Truncate(0))
	n, err := buf.Write([]byte{4, 5})

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, 2, n)
	assert.Exactly(t, 2, buf.Offset())
	assert.Exactly(t, 2, buf.Len())
	assert.Exactly(t, 8, buf.Cap())
	assert.Exactly(t, []byte{4, 5}, buf.buf)
	assert.NoError(t, buf.Close())
}

func Test_Buffer_Truncate_EdgeCaseWhenSizeEqualsLength(t *testing.T) {
	// --- Given ---
	buf, err := With([]byte{0, 1, 2, 3}, Append)
	require.NoError(t, err)

	// --- When ---
	assert.NoError(t, buf.Truncate(4))
	n, err := buf.Write([]byte{4, 5})

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, 2, n)
	assert.Exactly(t, 6, buf.Offset())
	assert.Exactly(t, 6, buf.Len())
	assert.Exactly(t, 6, buf.Cap())
	assert.Exactly(t, []byte{0, 1, 2, 3, 4, 5}, buf.buf)
	assert.NoError(t, buf.Close())
}

func Test_Buffer_Truncate_Error(t *testing.T) {
	// --- Given ---
	buf, err := With([]byte{0, 1, 2}, Append)
	require.NoError(t, err)

	// --- When ---
	err = buf.Truncate(-1)

	// --- Then ---
	assert.ErrorIs(t, err, os.ErrInvalid)
}
