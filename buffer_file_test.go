package flexbuf

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Tests in this file test Buffer behaves the same way as os.File.

// comFileBuffer is an interface common to os.File and Buffer.
type comFileBuffer interface {
	io.Seeker
	io.Reader
	io.ReaderAt
	io.Closer
	io.ReaderFrom
	io.Writer
	io.WriterAt
	io.StringWriter

	Truncate(size int64) error
}

func Test_File_ReadFrom_ToEmpty(t *testing.T) {
	tstFn := func(t *testing.T, buf comFileBuffer) {
		// --- Given ---
		src := bytes.NewBuffer(bytes.Repeat([]byte{1, 2}, 500))

		// --- When ---
		n, err := buf.ReadFrom(src)

		// --- Then ---
		assert.NoError(t, err)
		assert.Exactly(t, int64(1000), n)
		assert.Exactly(t, int64(1000), mustCurrOffset(buf))
		assert.Exactly(t, bytes.Repeat([]byte{1, 2}, 500), mustReadWhole(buf))
		assert.NoError(t, buf.Close())
	}

	// Setup file.
	tstFn(t, tmpFile(t))

	// Setup buffer.
	tstFn(t, &Buffer{})
}

func Test_File_ReadFrom_ToFull(t *testing.T) {
	testFn := func(t *testing.T, buf comFileBuffer) {
		// --- Given ---
		src := bytes.NewBuffer([]byte{3, 4, 5})

		// --- When ---
		n, err := buf.ReadFrom(src)

		// --- Then ---
		assert.NoError(t, err)
		assert.Exactly(t, int64(3), n)
		assert.Exactly(t, int64(6), mustCurrOffset(buf))
		want := []byte{0, 1, 2, 3, 4, 5}
		assert.Exactly(t, want, mustReadWhole(buf))
		assert.NoError(t, buf.Close())
	}

	// Setup file.
	testFn(t, tmpFileData(t, os.O_RDWR|os.O_APPEND, []byte{0, 1, 2}))

	// Setup buffer.
	testFn(t, With([]byte{0, 1, 2}, Append))
}

func Test_File_Write_OverrideAndExtend(t *testing.T) {
	testFn := func(t *testing.T, buf comFileBuffer) {
		// --- Given ---
		data := bytes.Repeat([]byte{0, 1}, 500)
		mustSeek(buf, 1, io.SeekStart)

		// --- When ---
		n, err := buf.Write(data)

		// --- Then ---
		assert.NoError(t, err)
		assert.Exactly(t, 1000, n)
		assert.Exactly(t, int64(1001), mustCurrOffset(buf))
		want := append([]byte{0}, data...)
		assert.Exactly(t, want, mustReadWhole(buf))
		assert.NoError(t, buf.Close())
	}

	// Setup file.
	testFn(t, tmpFileData(t, os.O_RDWR, []byte{0, 1, 2}))

	// Setup buffer.
	testFn(t, With([]byte{0, 1, 2}))
}

func Test_File_Write_append(t *testing.T) {
	testFn := func(t *testing.T, buf comFileBuffer) {
		// --- When ---
		n, err := buf.Write([]byte{3, 4, 5})

		// --- Then ---
		assert.NoError(t, err)
		assert.Exactly(t, 3, n)
		assert.Exactly(t, int64(6), mustCurrOffset(buf))
		assert.Exactly(t, []byte{0, 1, 2, 3, 4, 5}, mustReadWhole(buf))
		assert.NoError(t, buf.Close())
	}

	// Initial buffer value.
	in := []byte{0, 1, 2}

	// Setup file.
	fil := tmpFileData(t, os.O_RDWR|os.O_APPEND, in)
	testFn(t, fil)

	// Setup buffer.
	buf := With(in, Append)
	testFn(t, buf)
}

func Test_File_Write_override_and_extend(t *testing.T) {
	testFn := func(t *testing.T, buf comFileBuffer) {
		// --- When ---
		mustSeek(buf, 1, io.SeekStart)
		n, err := buf.Write([]byte{3, 4, 5})

		// --- Then ---
		assert.NoError(t, err)
		assert.Exactly(t, 3, n)
		assert.Exactly(t, int64(4), mustCurrOffset(buf))
		assert.Exactly(t, []byte{0, 3, 4, 5}, mustReadWhole(buf))
		assert.NoError(t, buf.Close())
	}

	// Initial buffer value.
	in := []byte{0, 1, 2}

	// Setup file.
	fil := tmpFileData(t, os.O_RDWR, in)
	testFn(t, fil)

	// Setup buffer.
	buf := With(in)
	testFn(t, buf)
}

func Test_File_Write_override_tail(t *testing.T) {
	testFn := func(t *testing.T, buf comFileBuffer) {
		// --- When ---
		mustSeek(buf, 1, io.SeekStart)
		n, err := buf.Write([]byte{3, 4})

		// --- Then ---
		assert.NoError(t, err)
		assert.Exactly(t, 2, n)
		assert.Exactly(t, int64(3), mustCurrOffset(buf))
		assert.Exactly(t, []byte{0, 3, 4}, mustReadWhole(buf))
		assert.NoError(t, buf.Close())
	}

	// Initial buffer value.
	in := []byte{0, 1, 2}

	// Setup file.
	fil := tmpFileData(t, os.O_RDWR, in)
	testFn(t, fil)

	// Setup buffer.
	buf := With(in)
	testFn(t, buf)
}

func Test_File_Write_override_middle(t *testing.T) {
	testFn := func(t *testing.T, buf comFileBuffer) {
		// --- When ---
		mustSeek(buf, 1, io.SeekStart)
		n, err := buf.Write([]byte{4, 5})

		// --- Then ---
		assert.NoError(t, err)
		assert.Exactly(t, 2, n)
		assert.Exactly(t, int64(3), mustCurrOffset(buf))
		assert.Exactly(t, []byte{0, 4, 5, 3}, mustReadWhole(buf))
		assert.NoError(t, buf.Close())
	}

	// Initial buffer value.
	in := []byte{0, 1, 2, 3}

	// Setup file.
	fil := tmpFileData(t, os.O_RDWR, in)
	testFn(t, fil)

	// Setup buffer.
	buf := With(in)
	testFn(t, buf)
}

func Test_File_WriteAt_ZeroValue(t *testing.T) {
	testFn := func(t *testing.T, buf comFileBuffer) {
		// --- When ---
		n, err := buf.WriteAt([]byte{0, 1, 2}, 0)

		// --- Then ---
		assert.NoError(t, err)
		assert.Exactly(t, 3, n)
		assert.Exactly(t, int64(0), mustCurrOffset(buf))
		want := []byte{0, 1, 2}
		assert.Exactly(t, want, mustReadWhole(buf))
		assert.NoError(t, buf.Close())
	}

	// Setup file.
	fil := tmpFile(t)
	testFn(t, fil)

	// Setup buffer.
	buf := &Buffer{}
	testFn(t, buf)
}

func Test_File_WriteAt_OverrideAndExtend(t *testing.T) {
	testFn := func(t *testing.T, buf comFileBuffer) {
		// --- When ---
		data := bytes.Repeat([]byte{0, 1}, 500)
		n, err := buf.WriteAt(data, 1)

		// --- Then ---
		assert.NoError(t, err)
		assert.Exactly(t, 1000, n)
		assert.Exactly(t, int64(0), mustCurrOffset(buf))
		want := append([]byte{0}, bytes.Repeat([]byte{0, 1}, 500)...)
		assert.Exactly(t, want, mustReadWhole(buf))
		assert.NoError(t, buf.Close())
	}

	// Initial buffer value.
	in := []byte{0, 1, 2}

	// Setup file.
	fil := tmpFileData(t, os.O_RDWR, in)
	testFn(t, fil)

	// Setup buffer.
	buf := With(in)
	testFn(t, buf)
}

func Test_File_WriteAt_BeyondCap(t *testing.T) {
	testFn := func(t *testing.T, buf comFileBuffer) {
		// --- When ---
		n, err := buf.WriteAt([]byte{3, 4, 5}, 1000)

		// --- Then ---
		assert.NoError(t, err)
		assert.Exactly(t, 3, n)
		assert.Exactly(t, int64(0), mustCurrOffset(buf))
		want := append([]byte{0, 1, 2}, bytes.Repeat([]byte{0}, 997)...)
		want = append(want, []byte{3, 4, 5}...)
		assert.Exactly(t, want, mustReadWhole(buf))
		assert.NoError(t, buf.Close())
	}

	// Initial buffer value.
	in := []byte{0, 1, 2}

	// Setup file.
	fil := tmpFileData(t, os.O_RDWR, in)
	testFn(t, fil)

	// Setup buffer.
	buf := With(in)
	testFn(t, buf)
}

func Test_File_WriteAt_append(t *testing.T) {
	testFn := func(t *testing.T, buf comFileBuffer) {
		// --- When ---
		n, err := buf.WriteAt([]byte{3, 4, 5}, 3)

		// --- Then ---
		assert.NoError(t, err)
		assert.Exactly(t, 3, n)
		assert.Exactly(t, int64(0), mustCurrOffset(buf))
		assert.Exactly(t, []byte{0, 1, 2, 3, 4, 5}, mustReadWhole(buf))
		assert.NoError(t, buf.Close())
	}

	// Initial buffer value.
	in := []byte{0, 1, 2}

	// Setup file.
	fil := tmpFileData(t, os.O_RDWR, in)
	testFn(t, fil)

	// Setup buffer.
	buf := With(in)
	testFn(t, buf)
}

func Test_File_WriteAt_override_and_extend(t *testing.T) {
	testFn := func(t *testing.T, buf comFileBuffer) {
		// --- When ---
		mustSeek(buf, 2, io.SeekStart)
		n, err := buf.WriteAt([]byte{3, 4, 5}, 1)

		// --- Then ---
		assert.NoError(t, err)
		assert.Exactly(t, 3, n)
		assert.Exactly(t, int64(2), mustCurrOffset(buf))
		assert.Exactly(t, []byte{0, 3, 4, 5}, mustReadWhole(buf))
		assert.NoError(t, buf.Close())
	}

	// Initial buffer value.
	in := []byte{0, 1, 2}

	// Setup file.
	fil := tmpFileData(t, os.O_RDWR, in)
	testFn(t, fil)

	// Setup buffer.
	buf := With(in)
	testFn(t, buf)
}

func Test_File_WriteAt_override_tail(t *testing.T) {
	testFn := func(t *testing.T, buf comFileBuffer) {
		// --- When ---
		mustSeek(buf, 2, io.SeekStart)
		n, err := buf.WriteAt([]byte{3, 4}, 1)

		// --- Then ---
		assert.NoError(t, err)
		assert.Exactly(t, 2, n)
		assert.Exactly(t, int64(2), mustCurrOffset(buf))
		assert.Exactly(t, []byte{0, 3, 4}, mustReadWhole(buf))
		assert.NoError(t, buf.Close())
	}

	// Initial buffer value.
	in := []byte{0, 1, 2}

	// Setup file.
	fil := tmpFileData(t, os.O_RDWR, in)
	testFn(t, fil)

	// Setup buffer.
	buf := With(in)
	testFn(t, buf)
}

func Test_File_WriteAt_override_middle(t *testing.T) {
	testFn := func(t *testing.T, buf comFileBuffer) {
		// --- When ---
		mustSeek(buf, 2, io.SeekStart)
		n, err := buf.WriteAt([]byte{4, 5}, 1)

		// --- Then ---
		assert.NoError(t, err)
		assert.Exactly(t, 2, n)
		assert.Exactly(t, int64(2), mustCurrOffset(buf))
		assert.Exactly(t, []byte{0, 4, 5, 3}, mustReadWhole(buf))
		assert.NoError(t, buf.Close())
	}

	// Initial buffer value.
	in := []byte{0, 1, 2, 3}

	// Setup file.
	fil := tmpFileData(t, os.O_RDWR, in)
	testFn(t, fil)

	// Setup buffer.
	buf := With(in)
	testFn(t, buf)
}

func Test_File_WriteAt_write_at_offset_beyond_cap(t *testing.T) {
	testFn := func(t *testing.T, buf comFileBuffer) {
		// --- When ---
		mustSeek(buf, 0, io.SeekStart)
		n, err := buf.WriteAt([]byte{1, 2}, 8)

		// --- Then ---
		assert.NoError(t, err)
		assert.Exactly(t, 2, n)
		assert.Exactly(t, int64(0), mustCurrOffset(buf))
		assert.Exactly(t, []byte{0, 0, 0, 0, 0, 0, 0, 0, 1, 2}, mustReadWhole(buf))
		assert.NoError(t, buf.Close())
	}

	// Initial buffer value.
	in := make([]byte, 3, 6)

	// Setup file.
	fil := tmpFileData(t, os.O_RDWR, in)
	testFn(t, fil)

	// Setup buffer.
	buf := With(in)
	testFn(t, buf)
}

func Test_File_WriteAt_write_at_offset_beyond_cap_offset_close_to_len(t *testing.T) {
	testFn := func(t *testing.T, buf comFileBuffer) {
		// --- When ---
		mustSeek(buf, 4, io.SeekStart)
		n, err := buf.WriteAt([]byte{1, 2}, 8)

		// --- Then ---
		assert.NoError(t, err)
		assert.Exactly(t, 2, n)
		assert.Exactly(t, int64(4), mustCurrOffset(buf))
		assert.Exactly(t, []byte{0, 0, 0, 0, 0, 0, 0, 0, 1, 2}, mustReadWhole(buf))
		assert.NoError(t, buf.Close())
	}

	// Initial buffer value.
	in := make([]byte, 5, 7)

	// Setup file.
	fil := tmpFileData(t, os.O_RDWR, in)
	testFn(t, fil)

	// Setup buffer.
	buf := With(in)
	testFn(t, buf)
}

func Test_File_WriteString(t *testing.T) {
	testFn := func(t *testing.T, buf comFileBuffer) {
		// --- Given ---
		mustSeek(buf, 1, io.SeekStart)

		// --- When ---
		n, err := buf.WriteString("abc")

		// --- Then ---
		assert.NoError(t, err)
		assert.Exactly(t, 3, n)
		assert.Exactly(t, []byte{0, 0x61, 0x62, 0x63}, mustReadWhole(buf))
		assert.Exactly(t, int64(4), mustCurrOffset(buf))
	}

	// Initial buffer value.
	in := []byte{0, 1, 2}

	// Setup file.
	fil := tmpFileData(t, os.O_RDWR, in)
	testFn(t, fil)

	// Setup buffer.
	buf := With(in)
	testFn(t, buf)

}

func Test_File_Read_ZeroValue(t *testing.T) {
	testFn := func(t *testing.T, buf comFileBuffer) {
		// --- When ---
		dst := make([]byte, 3)
		n, err := buf.Read(dst)

		// --- Then ---
		assert.ErrorIs(t, err, io.EOF)
		assert.Exactly(t, 0, n)
		assert.Exactly(t, int64(0), mustCurrOffset(buf))
		want := []byte{0, 0, 0}
		assert.Exactly(t, want, dst)
		assert.NoError(t, buf.Close())
	}

	// Setup file.
	fil := tmpFile(t)
	testFn(t, fil)

	// Setup buffer.
	buf := &Buffer{}
	testFn(t, buf)
}

func Test_File_Read_WithSmallBuffer(t *testing.T) {
	testFn := func(t *testing.T, buf comFileBuffer) {
		dst := make([]byte, 3)

		// --- Then ---

		// First read.
		n, err := buf.Read(dst)

		assert.NoError(t, err)
		assert.Exactly(t, 3, n)
		assert.Exactly(t, int64(3), mustCurrOffset(buf))
		want := []byte{0, 1, 2}
		assert.Exactly(t, want, dst)

		// Second read.
		n, err = buf.Read(dst)

		assert.NoError(t, err)
		assert.Exactly(t, 2, n)
		assert.Exactly(t, int64(5), mustCurrOffset(buf))
		want = []byte{3, 4, 2}
		assert.Exactly(t, want, dst)

		// Third read.
		n, err = buf.Read(dst)

		assert.ErrorIs(t, err, io.EOF)
		assert.Exactly(t, 0, n)
		assert.Exactly(t, int64(5), mustCurrOffset(buf))
		want = []byte{3, 4, 2}
		assert.Exactly(t, want, dst)

		assert.NoError(t, buf.Close())
	}

	// Initial buffer value.
	in := []byte{0, 1, 2, 3, 4}

	// Setup file.
	fil := tmpFileData(t, os.O_RDONLY, in)
	testFn(t, fil)

	// Setup buffer.
	buf := With(in)
	testFn(t, buf)
}

func Test_File_Read_BeyondLen(t *testing.T) {
	testFn := func(t *testing.T, buf comFileBuffer) {
		mustSeek(buf, 5, io.SeekStart)

		// --- When ---
		dst := make([]byte, 3)
		n, err := buf.Read(dst)

		// --- Then ---
		assert.ErrorIs(t, err, io.EOF)
		assert.Exactly(t, 0, n)
		assert.Exactly(t, int64(5), mustCurrOffset(buf))
		want := []byte{0, 0, 0}
		assert.Exactly(t, want, dst)
		assert.NoError(t, buf.Close())
	}

	// Initial buffer value.
	in := []byte{0, 1, 2}

	// Setup file.
	fil := tmpFileData(t, os.O_RDWR, in)
	testFn(t, fil)

	// Setup buffer.
	buf := With(in)
	testFn(t, buf)
}

func Test_File_Read_BigBuffer(t *testing.T) {
	testFn := func(t *testing.T, buf comFileBuffer) {
		// --- When ---
		dst := make([]byte, 6)
		n, err := buf.Read(dst)

		// --- Then ---
		assert.NoError(t, err)
		assert.Exactly(t, 3, n)
		assert.Exactly(t, int64(3), mustCurrOffset(buf))
		want := []byte{0, 1, 2, 0, 0, 0}
		assert.Exactly(t, want, dst)
		assert.NoError(t, buf.Close())
	}

	// Initial buffer value.
	in := []byte{0, 1, 2}

	// Setup file.
	fil := tmpFileData(t, os.O_RDWR, in)
	testFn(t, fil)

	// Setup buffer.
	buf := With(in)
	testFn(t, buf)
}

func Test_File_Read_read_all(t *testing.T) {
	testFn := func(t *testing.T, buf comFileBuffer) {
		dst := make([]byte, 3, 4)

		// --- When ---
		n, err := buf.Read(dst)

		// --- Then ---
		assert.NoError(t, err)
		assert.Exactly(t, 3, n)
		assert.Exactly(t, int64(3), mustCurrOffset(buf))
		assert.Exactly(t, []byte{0, 1, 2}, dst)
		assert.NoError(t, buf.Close())
	}

	// Initial buffer value.
	in := []byte{0, 1, 2}

	// Setup file.
	fil := tmpFileData(t, os.O_RDWR, in)
	testFn(t, fil)

	// Setup buffer.
	buf := With(in)
	testFn(t, buf)
}

func Test_File_Read_read_head(t *testing.T) {
	testFn := func(t *testing.T, buf comFileBuffer) {
		dst := make([]byte, 3, 4)

		// --- When ---
		n, err := buf.Read(dst)

		// --- Then ---
		assert.NoError(t, err)
		assert.Exactly(t, 3, n)
		assert.Exactly(t, int64(3), mustCurrOffset(buf))
		assert.Exactly(t, []byte{0, 1, 2}, dst)
		assert.NoError(t, buf.Close())
	}

	// Initial buffer value.
	in := []byte{0, 1, 2, 4}

	// Setup file.
	fil := tmpFileData(t, os.O_RDWR, in)
	testFn(t, fil)

	// Setup buffer.
	buf := With(in)
	testFn(t, buf)
}

func Test_File_Read_read_tail(t *testing.T) {
	testFn := func(t *testing.T, buf comFileBuffer) {
		mustSeek(buf, 1, io.SeekStart)
		dst := make([]byte, 2, 3)

		// --- When ---
		n, err := buf.Read(dst)

		// --- Then ---
		assert.NoError(t, err)
		assert.Exactly(t, 2, n)
		assert.Exactly(t, int64(3), mustCurrOffset(buf))
		assert.Exactly(t, []byte{1, 2}, dst)
		assert.NoError(t, buf.Close())
	}

	// Initial buffer value.
	in := []byte{0, 1, 2}

	// Setup file.
	fil := tmpFileData(t, os.O_RDWR, in)
	testFn(t, fil)

	// Setup buffer.
	buf := With(in)
	testFn(t, buf)
}

func Test_File_ReadAt_BeyondLen(t *testing.T) {
	testFn := func(t *testing.T, buf comFileBuffer) {
		// --- When ---
		dst := make([]byte, 4)
		n, err := buf.ReadAt(dst, 6)

		// --- Then ---
		assert.ErrorIs(t, err, io.EOF)
		assert.Exactly(t, 0, n)
		assert.Exactly(t, int64(0), mustCurrOffset(buf))
		want := []byte{0, 0, 0, 0}
		assert.Exactly(t, want, dst)
		assert.NoError(t, buf.Close())
	}

	// Initial buffer value.
	in := []byte{0, 1, 2}

	// Setup file.
	fil := tmpFileData(t, os.O_RDWR, in)
	testFn(t, fil)

	// Setup buffer.
	buf := With(in)
	testFn(t, buf)
}

func Test_File_ReadAt_BigBuffer(t *testing.T) {
	testFn := func(t *testing.T, buf comFileBuffer) {
		mustSeek(buf, 1, io.SeekStart)
		dst := make([]byte, 4)

		// --- When ---
		n, err := buf.ReadAt(dst, 0)

		// --- Then ---
		assert.ErrorIs(t, err, io.EOF)
		assert.Exactly(t, 3, n)
		assert.Exactly(t, int64(1), mustCurrOffset(buf))
		want := []byte{0, 1, 2, 0}
		assert.Exactly(t, want, dst)
		assert.NoError(t, buf.Close())
	}

	// Initial buffer value.
	in := []byte{0, 1, 2}

	// Setup file.
	fil := tmpFileData(t, os.O_RDWR, in)
	testFn(t, fil)

	// Setup buffer.
	buf := With(in)
	testFn(t, buf)
}

func Test_File_ReadAt_read_all(t *testing.T) {
	testFn := func(t *testing.T, buf comFileBuffer) {
		// --- Given ---
		mustSeek(buf, 1, io.SeekStart)
		dst := []byte{0, 1, 2}

		// --- When ---
		n, err := buf.ReadAt(dst, 0)

		// --- Then ---
		assert.NoError(t, err)
		assert.Exactly(t, 3, n)
		assert.Exactly(t, int64(1), mustCurrOffset(buf))
		assert.Exactly(t, []byte{0, 1, 2}, dst)
		assert.NoError(t, buf.Close())
	}

	// Initial buffer value.
	in := []byte{0, 1, 2}

	// Setup file.
	fil := tmpFileData(t, os.O_RDWR, in)
	testFn(t, fil)

	// Setup buffer.
	buf := With(in)
	testFn(t, buf)
}

func Test_File_ReadAt_read_head(t *testing.T) {
	testFn := func(t *testing.T, buf comFileBuffer) {
		// --- Given ---
		mustSeek(buf, 1, io.SeekStart)
		dst := make([]byte, 2, 3)

		// --- When ---
		n, err := buf.ReadAt(dst, 0)

		// --- Then ---
		assert.NoError(t, err)
		assert.Exactly(t, 2, n)
		assert.Exactly(t, int64(1), mustCurrOffset(buf))
		assert.Exactly(t, []byte{0, 1}, dst)
		assert.NoError(t, buf.Close())
	}

	// Initial buffer value.
	in := []byte{0, 1, 2}

	// Setup file.
	fil := tmpFileData(t, os.O_RDWR, in)
	testFn(t, fil)

	// Setup buffer.
	buf := With(in)
	testFn(t, buf)
}

func Test_File_ReadAt_read_tail(t *testing.T) {
	testFn := func(t *testing.T, buf comFileBuffer) {
		// --- Given ---
		mustSeek(buf, 2, io.SeekStart)
		dst := make([]byte, 2, 3)

		// --- When ---
		n, err := buf.ReadAt(dst, 1)

		// --- Then ---
		assert.NoError(t, err)
		assert.Exactly(t, 2, n)
		assert.Exactly(t, int64(2), mustCurrOffset(buf))
		assert.Exactly(t, []byte{1, 2}, dst)
		assert.NoError(t, buf.Close())
	}

	// Initial buffer value.
	in := []byte{0, 1, 2}

	// Setup file.
	fil := tmpFileData(t, os.O_RDWR, in)
	testFn(t, fil)

	// Setup buffer.
	buf := With(in)
	testFn(t, buf)
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
			fil := tmpFileData(t, os.O_RDWR, []byte{0, 1, 2, 3})
			mustSeek(fil, 1, io.SeekStart)

			buf := With([]byte{0, 1, 2, 3})
			mustSeek(buf, 1, io.SeekStart)

			// --- When ---
			filN, filErr := fil.Seek(tc.seek, tc.whence)
			bufN, bufErr := buf.Seek(tc.seek, tc.whence)

			// --- Then ---
			assert.NoError(t, filErr, "test %s", tc.testN)
			assert.NoError(t, bufErr, "test %s", tc.testN)

			assert.Exactly(t, tc.wantN, filN, "test %s", tc.testN)
			assert.Exactly(t, tc.wantN, bufN, "test %s", tc.testN)

			got, err := ioutil.ReadAll(fil)
			assert.NoError(t, err, "test %s", tc.testN)
			assert.Exactly(t, tc.wantD, got, "test %s", tc.testN)

			got, err = ioutil.ReadAll(buf)
			assert.NoError(t, err, "test %s", tc.testN)
			assert.Exactly(t, tc.wantD, got, "test %s", tc.testN)

			assert.NoError(t, fil.Close(), "test %s", tc.testN)
			assert.NoError(t, buf.Close(), "test %s", tc.testN)
		})
	}
}

func Test_File_Seek_BeyondLen(t *testing.T) {
	testFn := func(t *testing.T, buf comFileBuffer) {
		// --- When ---
		n, err := buf.Seek(5, io.SeekStart)

		// --- Then ---
		assert.NoError(t, err)
		assert.Exactly(t, int64(5), n)
	}

	// Initial buffer value.
	in := []byte{0, 1, 2}

	// Setup file.
	fil := tmpFileData(t, os.O_RDWR, in)
	testFn(t, fil)

	// Setup buffer.
	buf := With(in)
	testFn(t, buf)
}

func Test_File_Truncate_ToZero(t *testing.T) {
	testFn := func(t *testing.T, buf comFileBuffer) {
		// --- When ---
		err := buf.Truncate(0)

		// --- Then ---
		assert.NoError(t, err)
		assert.Exactly(t, int64(0), mustCurrOffset(buf))
		assert.Exactly(t, []byte{}, mustReadWhole(buf))
		assert.NoError(t, buf.Close())
	}

	// Initial buffer value.
	in := []byte{0, 1, 2, 3}

	// Setup file.
	fil := tmpFileData(t, os.O_RDWR, in)
	testFn(t, fil)

	// Setup buffer.
	buf := With(in)
	testFn(t, buf)
}

func Test_File_Truncate_ToOne(t *testing.T) {
	testFn := func(t *testing.T, buf comFileBuffer) {
		// --- When ---
		err := buf.Truncate(1)

		// --- Then ---
		assert.NoError(t, err)
		assert.Exactly(t, int64(0), mustCurrOffset(buf))
		assert.Exactly(t, []byte{0}, mustReadWhole(buf))
		assert.NoError(t, buf.Close())
	}

	// Initial buffer value.
	in := []byte{0, 1, 2, 3}

	// Setup file.
	fil := tmpFileData(t, os.O_RDWR, in)
	testFn(t, fil)

	// Setup buffer.
	buf := With(in)
	testFn(t, buf)
}

func Test_File_Truncate_ToZeroAndWrite(t *testing.T) {
	testFn := func(t *testing.T, buf comFileBuffer) {
		// --- When ---
		err := buf.Truncate(0)
		assert.NoError(t, err)

		n, err := buf.Write([]byte{4, 5})

		// --- Then ---
		assert.NoError(t, err)
		assert.Exactly(t, 2, n)
		assert.Exactly(t, int64(2), mustCurrOffset(buf))
		assert.Exactly(t, []byte{4, 5}, mustReadWhole(buf))
		assert.NoError(t, buf.Close())
	}

	// Initial buffer value.
	in := []byte{0, 1, 2, 3}

	// Setup file.
	fil := tmpFileData(t, os.O_RDWR, in)
	testFn(t, fil)

	// Setup buffer.
	buf := With(in)
	testFn(t, buf)
}

func Test_File_Truncate_BeyondLenAndWrite(t *testing.T) {
	testFn := func(t *testing.T, buf comFileBuffer) {
		mustSeek(buf, 1, io.SeekStart)

		// --- When ---
		assert.NoError(t, buf.Truncate(8))
		n, err := buf.Write([]byte{4, 5})

		// --- Then ---
		assert.NoError(t, err)
		assert.Exactly(t, 2, n)
		assert.Exactly(t, int64(10), mustCurrOffset(buf))
		assert.Exactly(t, []byte{0, 1, 2, 3, 0, 0, 0, 0, 4, 5}, mustReadWhole(buf))
		assert.NoError(t, buf.Close())
	}

	// Initial buffer value.
	in := []byte{0, 1, 2, 3}

	// Setup file.
	fil := tmpFileData(t, os.O_RDWR|os.O_APPEND, in)
	testFn(t, fil)

	// Setup buffer.
	buf := With(in, Append)
	testFn(t, buf)
}

func Test_File_Truncate_BeyondCapAndWrite(t *testing.T) {
	testFn := func(t *testing.T, buf comFileBuffer) {
		// --- When ---
		assert.NoError(t, buf.Truncate(10))
		n, err := buf.Write([]byte{4, 5})

		// --- Then ---
		assert.NoError(t, err)
		assert.Exactly(t, 2, n)
		assert.Exactly(t, int64(12), mustCurrOffset(buf))
		want := []byte{0, 1, 2, 3, 0, 0, 0, 0, 0, 0, 4, 5}
		assert.Exactly(t, want, mustReadWhole(buf))
		assert.NoError(t, buf.Close())
	}

	// Initial buffer value.
	in := make([]byte, 4, 8)
	in[0] = 0
	in[1] = 1
	in[2] = 2
	in[3] = 3

	// Setup file.
	fil := tmpFileData(t, os.O_RDWR|os.O_APPEND, in)
	testFn(t, fil)

	// Setup buffer.
	buf := With(in, Append)
	testFn(t, buf)
}

func Test_File_Truncate_ExtendBeyondLenResetAndWrite(t *testing.T) {
	testFn := func(t *testing.T, buf comFileBuffer) {
		// --- When ---
		assert.NoError(t, buf.Truncate(8))
		assert.NoError(t, buf.Truncate(0))
		n, err := buf.Write([]byte{4, 5})

		// --- Then ---
		assert.NoError(t, err)
		assert.Exactly(t, 2, n)
		assert.Exactly(t, int64(2), mustCurrOffset(buf))
		assert.Exactly(t, []byte{4, 5}, mustReadWhole(buf))
		assert.NoError(t, buf.Close())
	}

	// Initial buffer value.
	in := []byte{0, 1, 2, 3}

	// Setup file.
	fil := tmpFileData(t, os.O_RDWR|os.O_APPEND, in)
	testFn(t, fil)

	// Setup buffer.
	buf := With(in, Append)
	testFn(t, buf)
}

func Test_File_Truncate_EdgeCaseWhenSizeEqualsLength(t *testing.T) {
	testFn := func(t *testing.T, buf comFileBuffer) {
		// --- When ---
		assert.NoError(t, buf.Truncate(4))
		n, err := buf.Write([]byte{4, 5})

		// --- Then ---
		assert.NoError(t, err)
		assert.Exactly(t, 2, n)
		assert.Exactly(t, int64(6), mustCurrOffset(buf))
		assert.Exactly(t, []byte{0, 1, 2, 3, 4, 5}, mustReadWhole(buf))
		assert.NoError(t, buf.Close())
	}

	// Initial buffer value.
	in := []byte{0, 1, 2, 3}

	// Setup file.
	fil := tmpFileData(t, os.O_RDWR|os.O_APPEND, in)
	testFn(t, fil)

	// Setup buffer.
	buf := With(in, Append)
	testFn(t, buf)
}
