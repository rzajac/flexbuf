package flexbuf

import (
	"bytes"
	"io"
	"io/ioutil"
	"math"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Tests in this file are to confirm some of the os.File behaviours
// so they can be matched by flexbuf.Buffer.

func Test_File_ReadFrom_ToEmpty(t *testing.T) {
	// --- Given ---
	src := bytes.NewBuffer(bytes.Repeat([]byte{1, 2}, 500))
	buf := tmpFile(t)

	// --- When ---
	n, err := buf.ReadFrom(src)

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, int64(1000), n)
	assert.Exactly(t, int64(1000), mustCurrOffset(buf))
	assert.Exactly(t, int64(1000), mustFileSize(buf))
	assert.Exactly(t, bytes.Repeat([]byte{1, 2}, 500), mustReadWhole(buf))
	assert.NoError(t, buf.Close())
}

func Test_File_ReadFrom_ToFull(t *testing.T) {
	// --- Given ---
	src := bytes.NewBuffer([]byte{3, 4, 5})
	buf := tmpFileData(t, os.O_RDWR|os.O_APPEND, []byte{0, 1, 2})

	// --- When ---
	n, err := buf.ReadFrom(src)

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, int64(3), n)
	assert.Exactly(t, int64(6), mustCurrOffset(buf))
	assert.Exactly(t, int64(6), mustFileSize(buf))
	want := []byte{0, 1, 2, 3, 4, 5}
	assert.Exactly(t, want, mustReadWhole(buf))
	assert.NoError(t, buf.Close())
}

func Test_File_Write_OverrideAndExtend(t *testing.T) {
	// --- Given ---
	data := bytes.Repeat([]byte{0, 1}, 500)
	buf := tmpFileData(t, os.O_RDWR, []byte{0, 1, 2})
	mustSeek(buf, 1, io.SeekStart)

	// --- When ---
	n, err := buf.Write(data)

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, 1000, n)
	assert.Exactly(t, int64(1001), mustCurrOffset(buf))
	assert.Exactly(t, int64(1001), mustFileSize(buf))
	want := append([]byte{0}, data...)
	assert.Exactly(t, want, mustReadWhole(buf))
	assert.NoError(t, buf.Close())
}

func Test_File_Write(t *testing.T) {
	const dontSeek = math.MinInt64

	tt := []struct {
		testN string

		init   []byte
		flag   int
		seek   int64
		src    []byte
		expN   int
		expOff int64
		expLen int64
		expBuf []byte
	}{
		{
			testN:  "append",
			init:   []byte{0, 1, 2},
			flag:   os.O_RDWR | os.O_APPEND,
			seek:   dontSeek,
			src:    []byte{3, 4, 5},
			expN:   3,
			expOff: 6,
			expLen: 6,
			expBuf: []byte{0, 1, 2, 3, 4, 5},
		},
		{
			testN:  "override and extend",
			init:   []byte{0, 1, 2},
			flag:   os.O_RDWR,
			seek:   1,
			src:    []byte{3, 4, 5},
			expN:   3,
			expOff: 4,
			expLen: 4,
			expBuf: []byte{0, 3, 4, 5},
		},
		{
			testN:  "override tail",
			init:   []byte{0, 1, 2},
			flag:   os.O_RDWR,
			seek:   1,
			src:    []byte{3, 4},
			expN:   2,
			expOff: 3,
			expLen: 3,
			expBuf: []byte{0, 3, 4},
		},
		{
			testN:  "override middle",
			init:   []byte{0, 1, 2, 3},
			flag:   os.O_RDWR,
			seek:   1,
			src:    []byte{4, 5},
			expN:   2,
			expOff: 3,
			expLen: 4,
			expBuf: []byte{0, 4, 5, 3},
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			buf := tmpFileData(t, tc.flag, tc.init)
			if tc.seek != dontSeek {
				mustSeek(buf, tc.seek, io.SeekStart)
			}

			// --- When ---
			n, err := buf.Write(tc.src)

			// --- Then ---
			assert.NoError(t, err, "test %s", tc.testN)
			assert.Exactly(t, tc.expN, n, "test %s", tc.testN)
			assert.Exactly(t, tc.expOff, mustCurrOffset(buf), "test %s", tc.testN)
			assert.Exactly(t, tc.expLen, mustFileSize(buf), "test %s", tc.testN)
			assert.Exactly(t, tc.expBuf, mustReadWhole(buf), "test %s", tc.testN)
			assert.NoError(t, buf.Close(), "test %s", tc.testN)
		})
	}
}

func Test_File_WriteAt_ZeroValue(t *testing.T) {
	// --- Given ---
	buf := tmpFile(t)

	// --- When ---
	n, err := buf.WriteAt([]byte{0, 1, 2}, 0)

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, 3, n)
	assert.Exactly(t, int64(0), mustCurrOffset(buf))
	assert.Exactly(t, int64(3), mustFileSize(buf))
	want := []byte{0, 1, 2}
	assert.Exactly(t, want, mustReadWhole(buf))
	assert.NoError(t, buf.Close())
}

func Test_File_WriteAt_OverrideAndExtend(t *testing.T) {
	// --- Given ---
	buf := tmpFileData(t, os.O_RDWR, []byte{0, 1, 2})

	// --- When ---
	data := bytes.Repeat([]byte{0, 1}, 500)
	n, err := buf.WriteAt(data, 1)

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, 1000, n)
	assert.Exactly(t, int64(0), mustCurrOffset(buf))
	assert.Exactly(t, int64(1001), mustFileSize(buf))
	want := append([]byte{0}, bytes.Repeat([]byte{0, 1}, 500)...)
	assert.Exactly(t, want, mustReadWhole(buf))
	assert.NoError(t, buf.Close())
}

func Test_File_WriteAt_BeyondCap(t *testing.T) {
	// --- Given ---
	buf := tmpFileData(t, os.O_RDWR, []byte{0, 1, 2})

	// --- When ---
	n, err := buf.WriteAt([]byte{3, 4, 5}, 1000)

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, 3, n)
	assert.Exactly(t, int64(0), mustCurrOffset(buf))
	assert.Exactly(t, int64(1003), mustFileSize(buf))
	want := append([]byte{0, 1, 2}, bytes.Repeat([]byte{0}, 997)...)
	want = append(want, []byte{3, 4, 5}...)
	assert.Exactly(t, want, mustReadWhole(buf))
	assert.NoError(t, buf.Close())
}

func Test_File_WriteAt(t *testing.T) {
	const dontSeek = math.MinInt64

	tt := []struct {
		testN string

		init   []byte
		flag   int
		seek   int64
		src    []byte
		off    int64
		expN   int
		expOff int64
		expLen int64
		expBuf []byte
	}{
		{
			testN:  "append",
			init:   []byte{0, 1, 2},
			flag:   os.O_RDWR,
			seek:   dontSeek,
			src:    []byte{3, 4, 5},
			off:    3,
			expN:   3,
			expOff: 0,
			expLen: 6,
			expBuf: []byte{0, 1, 2, 3, 4, 5},
		},
		{
			testN:  "override and extend",
			init:   []byte{0, 1, 2},
			flag:   os.O_RDWR,
			seek:   2,
			src:    []byte{3, 4, 5},
			off:    1,
			expN:   3,
			expOff: 2,
			expLen: 4,
			expBuf: []byte{0, 3, 4, 5},
		},
		{
			testN:  "override tail",
			init:   []byte{0, 1, 2},
			flag:   os.O_RDWR,
			seek:   2,
			src:    []byte{3, 4},
			off:    1,
			expN:   2,
			expOff: 2,
			expLen: 3,
			expBuf: []byte{0, 3, 4},
		},
		{
			testN:  "override middle",
			init:   []byte{0, 1, 2, 3},
			flag:   os.O_RDWR,
			seek:   2,
			src:    []byte{4, 5},
			off:    1,
			expN:   2,
			expOff: 2,
			expLen: 4,
			expBuf: []byte{0, 4, 5, 3},
		},
		{
			testN:  "write at offset beyond cap",
			init:   make([]byte, 3, 6),
			flag:   os.O_RDWR,
			seek:   0,
			src:    []byte{1, 2},
			off:    8,
			expN:   2,
			expOff: 0,
			expLen: 10,
			expBuf: []byte{0, 0, 0, 0, 0, 0, 0, 0, 1, 2},
		},
		{
			testN:  "write at offset beyond cap, offset close to len",
			init:   make([]byte, 5, 7),
			flag:   os.O_RDWR,
			seek:   4,
			src:    []byte{1, 2},
			off:    8,
			expN:   2,
			expOff: 4,
			expLen: 10,
			expBuf: []byte{0, 0, 0, 0, 0, 0, 0, 0, 1, 2},
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			buf := tmpFileData(t, tc.flag, tc.init)
			if tc.seek != dontSeek {
				mustSeek(buf, tc.seek, io.SeekStart)
			}

			// --- When ---
			n, err := buf.WriteAt(tc.src, tc.off)

			// --- Then ---
			assert.NoError(t, err, "test %s", tc.testN)
			assert.Exactly(t, tc.expN, n, "test %s", tc.testN)
			assert.Exactly(t, tc.expOff, mustCurrOffset(buf), "test %s", tc.testN)
			assert.Exactly(t, tc.expLen, mustFileSize(buf), "test %s", tc.testN)
			assert.Exactly(t, tc.expBuf, mustReadWhole(buf), "test %s", tc.testN)
			assert.NoError(t, buf.Close(), "test %s", tc.testN)
		})
	}
}

func Test_File_WriteString(t *testing.T) {
	// --- Given ---
	buf := tmpFileData(t, os.O_RDWR, []byte{0, 1, 2})
	mustSeek(buf, 1, io.SeekStart)

	// --- When ---
	n, err := buf.WriteString("abc")

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, 3, n)
	assert.Exactly(t, []byte{0, 0x61, 0x62, 0x63}, mustReadWhole(buf))
	assert.Exactly(t, int64(4), mustCurrOffset(buf))
}

func Test_File_Read_ZeroValue(t *testing.T) {
	// --- Given ---
	buf := tmpFile(t)

	// --- When ---
	dst := make([]byte, 3)
	n, err := buf.Read(dst)

	// --- Then ---
	assert.ErrorIs(t, err, io.EOF)
	assert.Exactly(t, 0, n)
	assert.Exactly(t, int64(0), mustCurrOffset(buf))
	assert.Exactly(t, int64(0), mustFileSize(buf))
	want := []byte{0, 0, 0}
	assert.Exactly(t, want, dst)
	assert.NoError(t, buf.Close())
}

func Test_File_Read_WithSmallBuffer(t *testing.T) {
	// --- Given ---
	buf := tmpFileData(t, os.O_RDONLY, []byte{0, 1, 2, 3, 4})
	dst := make([]byte, 3)

	// --- Then ---

	// First read.
	n, err := buf.Read(dst)

	assert.NoError(t, err)
	assert.Exactly(t, 3, n)
	assert.Exactly(t, int64(3), mustCurrOffset(buf))
	assert.Exactly(t, int64(5), mustFileSize(buf))
	want := []byte{0, 1, 2}
	assert.Exactly(t, want, dst)

	// Second read.
	n, err = buf.Read(dst)

	assert.NoError(t, err)
	assert.Exactly(t, 2, n)
	assert.Exactly(t, int64(5), mustCurrOffset(buf))
	assert.Exactly(t, int64(5), mustFileSize(buf))
	want = []byte{3, 4, 2}
	assert.Exactly(t, want, dst)

	// Third read.
	n, err = buf.Read(dst)

	assert.ErrorIs(t, err, io.EOF)
	assert.Exactly(t, 0, n)
	assert.Exactly(t, int64(5), mustCurrOffset(buf))
	assert.Exactly(t, int64(5), mustFileSize(buf))
	want = []byte{3, 4, 2}
	assert.Exactly(t, want, dst)

	assert.NoError(t, buf.Close())
}

func Test_File_Read_BeyondLen(t *testing.T) {
	// --- Given ---
	buf := tmpFileData(t, os.O_RDWR, []byte{0, 1, 2})
	mustSeek(buf, 5, io.SeekStart)

	// --- When ---
	dst := make([]byte, 3)
	n, err := buf.Read(dst)

	// --- Then ---
	assert.ErrorIs(t, err, io.EOF)
	assert.Exactly(t, 0, n)
	assert.Exactly(t, int64(5), mustCurrOffset(buf))
	assert.Exactly(t, int64(3), mustFileSize(buf))
	want := []byte{0, 0, 0}
	assert.Exactly(t, want, dst)
	assert.NoError(t, buf.Close())
}

func Test_File_Read_BigBuffer(t *testing.T) {
	// --- Given ---
	buf := tmpFileData(t, os.O_RDWR, []byte{0, 1, 2})

	// --- When ---
	dst := make([]byte, 6)
	n, err := buf.Read(dst)

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, 3, n)
	assert.Exactly(t, int64(3), mustCurrOffset(buf))
	assert.Exactly(t, int64(3), mustFileSize(buf))
	want := []byte{0, 1, 2, 0, 0, 0}
	assert.Exactly(t, want, dst)
	assert.NoError(t, buf.Close())
}

func Test_File_Read(t *testing.T) {
	const dontSeek = math.MinInt64

	tt := []struct {
		testN string

		init   []byte
		flag   int
		seek   int64
		dst    []byte
		expN   int
		expOff int64
		expLen int64
		expDst []byte
	}{
		{
			testN:  "read all",
			init:   []byte{0, 1, 2},
			flag:   os.O_RDWR,
			seek:   dontSeek,
			dst:    make([]byte, 3, 3),
			expN:   3,
			expOff: 3,
			expLen: 3,
			expDst: []byte{0, 1, 2},
		},
		{
			testN:  "read head",
			init:   []byte{0, 1, 2},
			flag:   os.O_RDWR,
			seek:   dontSeek,
			dst:    make([]byte, 2, 3),
			expN:   2,
			expOff: 2,
			expLen: 3,
			expDst: []byte{0, 1},
		},
		{
			testN:  "read tail",
			init:   []byte{0, 1, 2},
			flag:   os.O_RDWR,
			seek:   1,
			dst:    make([]byte, 2, 3),
			expN:   2,
			expOff: 3,
			expLen: 3,
			expDst: []byte{1, 2},
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			buf := tmpFileData(t, tc.flag, tc.init)
			if tc.seek != dontSeek {
				mustSeek(buf, tc.seek, io.SeekStart)
			}

			// --- When ---
			n, err := buf.Read(tc.dst)

			// --- Then ---
			assert.NoError(t, err, "test %s", tc.testN)
			assert.Exactly(t, tc.expN, n, "test %s", tc.testN)
			assert.Exactly(t, tc.expOff, mustCurrOffset(buf), "test %s", tc.testN)
			assert.Exactly(t, tc.expLen, mustFileSize(buf), "test %s", tc.testN)
			assert.Exactly(t, tc.expDst, tc.dst, "test %s", tc.testN)
			assert.NoError(t, buf.Close(), "test %s", tc.testN)
		})
	}
}

func Test_File_ReadAt_BeyondLen(t *testing.T) {
	// --- Given ---
	buf := tmpFileData(t, os.O_RDWR, []byte{0, 1, 2})

	// --- When ---
	dst := make([]byte, 4)
	n, err := buf.ReadAt(dst, 6)

	// --- Then ---
	assert.ErrorIs(t, err, io.EOF)
	assert.Exactly(t, 0, n)
	assert.Exactly(t, int64(0), mustCurrOffset(buf))
	assert.Exactly(t, int64(3), mustFileSize(buf))
	want := []byte{0, 0, 0, 0}
	assert.Exactly(t, want, dst)
	assert.NoError(t, buf.Close())
}

func Test_File_ReadAt_BigBuffer(t *testing.T) {
	// --- Given ---
	buf := tmpFileData(t, os.O_RDWR, []byte{0, 1, 2})
	mustSeek(buf, 1, io.SeekStart)
	dst := make([]byte, 4)

	// --- When ---
	n, err := buf.ReadAt(dst, 0)

	// --- Then ---
	assert.ErrorIs(t, err, io.EOF)
	assert.Exactly(t, 3, n)
	assert.Exactly(t, int64(1), mustCurrOffset(buf))
	assert.Exactly(t, int64(3), mustFileSize(buf))
	want := []byte{0, 1, 2, 0}
	assert.Exactly(t, want, dst)
	assert.NoError(t, buf.Close())
}

func Test_File_ReadAt(t *testing.T) {
	const dontSeek = math.MinInt64

	tt := []struct {
		testN string

		init   []byte
		flag   int
		seek   int64
		dst    []byte
		off    int64
		expN   int
		expOff int64
		expLen int64
		expDst []byte
	}{
		{
			testN:  "read all",
			init:   []byte{0, 1, 2},
			flag:   os.O_RDWR,
			seek:   1,
			dst:    make([]byte, 3),
			off:    0,
			expN:   3,
			expOff: 1,
			expLen: 3,
			expDst: []byte{0, 1, 2},
		},
		{
			testN:  "read head",
			init:   []byte{0, 1, 2},
			flag:   os.O_RDWR,
			seek:   1,
			dst:    make([]byte, 2, 3),
			off:    0,
			expN:   2,
			expOff: 1,
			expLen: 3,
			expDst: []byte{0, 1},
		},
		{
			testN:  "read tail",
			init:   []byte{0, 1, 2},
			flag:   os.O_RDWR,
			seek:   2,
			dst:    make([]byte, 2, 3),
			off:    1,
			expN:   2,
			expOff: 2,
			expLen: 3,
			expDst: []byte{1, 2},
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			buf := tmpFileData(t, tc.flag, tc.init)
			if tc.seek != dontSeek {
				mustSeek(buf, tc.seek, io.SeekStart)
			}

			// --- When ---
			n, err := buf.ReadAt(tc.dst, tc.off)

			// --- Then ---
			assert.NoError(t, err, "test %s", tc.testN)
			assert.Exactly(t, tc.expN, n, "test %s", tc.testN)
			assert.Exactly(t, tc.expOff, mustCurrOffset(buf), "test %s", tc.testN)
			assert.Exactly(t, tc.expLen, mustFileSize(buf), "test %s", tc.testN)
			assert.Exactly(t, tc.expDst, tc.dst, "test %s", tc.testN)
			assert.NoError(t, buf.Close(), "test %s", tc.testN)
		})
	}
}

func Test_File_Seek(t *testing.T) {
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
			buf := tmpFileData(t, os.O_RDWR, []byte{0, 1, 2, 3})
			mustSeek(buf, 1, io.SeekStart)

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

func Test_File_Seek_NegativeFinalOffset(t *testing.T) {
	// --- Given ---
	buf := tmpFileData(t, os.O_RDWR, []byte{0, 1, 2})

	// --- When ---
	n, err := buf.Seek(-4, io.SeekEnd)

	// --- Then ---
	assert.IsType(t, err, &os.PathError{}) // File<>Buffer
	assert.Exactly(t, int64(0), n)
}

func Test_File_Seek_BeyondLen(t *testing.T) {
	// --- Given ---
	buf := tmpFileData(t, os.O_RDWR, []byte{0, 1, 2})

	// --- When ---
	n, err := buf.Seek(5, io.SeekStart)

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, int64(5), n)
}

func Test_File_Truncate_ToZero(t *testing.T) {
	// --- Given ---
	buf := tmpFileData(t, os.O_RDWR, []byte{0, 1, 2, 3})

	// --- When ---
	err := buf.Truncate(0)

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, int64(0), mustCurrOffset(buf))
	assert.Exactly(t, int64(0), mustFileSize(buf))
	assert.Exactly(t, []byte{}, mustReadWhole(buf))
	assert.NoError(t, buf.Close())
}

func Test_File_Truncate_ToOne(t *testing.T) {
	// --- Given ---
	buf := tmpFileData(t, os.O_RDWR, []byte{0, 1, 2, 3})

	// --- When ---
	err := buf.Truncate(1)

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, int64(0), mustCurrOffset(buf))
	assert.Exactly(t, int64(1), mustFileSize(buf))
	assert.Exactly(t, []byte{0}, mustReadWhole(buf))
	assert.NoError(t, buf.Close())
}

func Test_File_Truncate_ToZeroAndWrite(t *testing.T) {
	// --- Given ---
	buf := tmpFileData(t, os.O_RDWR, []byte{0, 1, 2, 3})

	// --- When ---
	err := buf.Truncate(0)
	assert.NoError(t, err)

	n, err := buf.Write([]byte{4, 5})

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, 2, n)
	assert.Exactly(t, int64(2), mustCurrOffset(buf))
	assert.Exactly(t, int64(2), mustFileSize(buf))
	assert.Exactly(t, []byte{4, 5}, mustReadWhole(buf))
	assert.NoError(t, buf.Close())
}

func Test_File_Truncate_BeyondLenAndWrite(t *testing.T) {
	// --- Given ---
	buf := tmpFileData(t, os.O_RDWR|os.O_APPEND, []byte{0, 1, 2, 3})
	mustSeek(buf, 1, io.SeekStart)

	// --- When ---
	assert.NoError(t, buf.Truncate(8))
	n, err := buf.Write([]byte{4, 5})

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, 2, n)
	assert.Exactly(t, int64(10), mustCurrOffset(buf))
	assert.Exactly(t, int64(10), mustFileSize(buf))
	assert.Exactly(t, []byte{0, 1, 2, 3, 0, 0, 0, 0, 4, 5}, mustReadWhole(buf))
	assert.NoError(t, buf.Close())
}

func Test_File_Truncate_BeyondCapAndWrite(t *testing.T) {
	// --- Given ---
	data := make([]byte, 4, 8)
	data[0] = 0
	data[1] = 1
	data[2] = 2
	data[3] = 3
	buf := tmpFileData(t, os.O_RDWR|os.O_APPEND, data)

	// --- When ---
	assert.NoError(t, buf.Truncate(10))
	n, err := buf.Write([]byte{4, 5})

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, 2, n)
	assert.Exactly(t, int64(12), mustCurrOffset(buf))
	assert.Exactly(t, int64(12), mustFileSize(buf))
	want := []byte{0, 1, 2, 3, 0, 0, 0, 0, 0, 0, 4, 5}
	assert.Exactly(t, want, mustReadWhole(buf))
	assert.NoError(t, buf.Close())
}

func Test_File_Truncate_ExtendBeyondLenResetAndWrite(t *testing.T) {
	// --- Given ---
	buf := tmpFileData(t, os.O_RDWR|os.O_APPEND, []byte{0, 1, 2, 3})

	// --- When ---
	assert.NoError(t, buf.Truncate(8))
	assert.NoError(t, buf.Truncate(0))
	n, err := buf.Write([]byte{4, 5})

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, 2, n)
	assert.Exactly(t, int64(2), mustCurrOffset(buf))
	assert.Exactly(t, int64(2), mustFileSize(buf))
	assert.Exactly(t, []byte{4, 5}, mustReadWhole(buf))
	assert.NoError(t, buf.Close())
}

func Test_File_Truncate_EdgeCaseWhenSizeEqualsLength(t *testing.T) {
	// --- Given ---
	buf := tmpFileData(t, os.O_RDWR|os.O_APPEND, []byte{0, 1, 2, 3})

	// --- When ---
	assert.NoError(t, buf.Truncate(4))
	n, err := buf.Write([]byte{4, 5})

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, 2, n)
	assert.Exactly(t, int64(6), mustCurrOffset(buf))
	assert.Exactly(t, int64(6), mustFileSize(buf))
	assert.Exactly(t, []byte{0, 1, 2, 3, 4, 5}, mustReadWhole(buf))
	assert.NoError(t, buf.Close())
}

func Test_File_Truncate_Error(t *testing.T) {
	// --- Given ---
	buf := tmpFileData(t, os.O_RDWR|os.O_APPEND, []byte{0, 1, 2})

	// --- When ---
	err := buf.Truncate(-1)

	// --- Then ---
	assert.IsType(t, err, &os.PathError{}) // File<>Buffer
}
