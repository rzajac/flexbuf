package flexbuf

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"testing"

	kit "github.com/rzajac/testkit"
	"github.com/stretchr/testify/assert"
)

// Tests in this file test Buffer behaves the same way as os.File.

// filer is an interface common to os.File and Buffer.
type filer interface {
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

func Test_File_ReadFrom_toEmpty(t *testing.T) {
	tt := []struct {
		testN string

		buf filer
	}{
		{"fil", kit.TempFile(t, t.TempDir(), "")},
		{"buf", &Buffer{}},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			src := bytes.Repeat([]byte{1, 2}, 500)

			// --- When ---
			n, err := tc.buf.ReadFrom(bytes.NewBuffer(src))

			// --- Then ---
			assert.NoError(t, err)
			assert.Exactly(t, int64(1000), n)
			assert.Exactly(t, int64(1000), kit.CurrOffset(t, tc.buf))
			assert.Exactly(t, src, kit.ReadAllFromStart(t, tc.buf))
			assert.NoError(t, tc.buf.Close())
		})
	}
}

func Test_File_ReadFrom_toFull(t *testing.T) {
	tt := []struct {
		testN string

		buf filer
	}{
		{"fil", TempFile(t, os.O_RDWR|os.O_APPEND, []byte{0, 1, 2})},
		{"buf", With([]byte{0, 1, 2}, Append)},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			src := bytes.NewBuffer([]byte{3, 4, 5})

			// --- When ---
			n, err := tc.buf.ReadFrom(src)

			// --- Then ---
			assert.NoError(t, err)
			assert.Exactly(t, int64(3), n)
			assert.Exactly(t, int64(6), kit.CurrOffset(t, tc.buf))
			want := []byte{0, 1, 2, 3, 4, 5}
			assert.Exactly(t, want, kit.ReadAllFromStart(t, tc.buf))
			assert.NoError(t, tc.buf.Close())
		})
	}
}

func Test_File_Write_append(t *testing.T) {
	tt := []struct {
		testN string

		buf filer
	}{
		{"fil", TempFile(t, os.O_RDWR|os.O_APPEND, []byte{0, 1, 2})},
		{"buf", With([]byte{0, 1, 2}, Append)},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			n, err := tc.buf.Write([]byte{3, 4, 5})

			// --- Then ---
			assert.NoError(t, err)
			assert.Exactly(t, 3, n)
			assert.Exactly(t, int64(6), kit.CurrOffset(t, tc.buf))

			exp := []byte{0, 1, 2, 3, 4, 5}
			assert.Exactly(t, exp, kit.ReadAllFromStart(t, tc.buf))
			assert.NoError(t, tc.buf.Close())
		})
	}
}

func Test_File_Write_overrideAndExtend(t *testing.T) {
	tt := []struct {
		testN string

		buf filer
	}{
		{"fil", TempFile(t, os.O_RDWR, []byte{0, 1, 2})},
		{"buf", With([]byte{0, 1, 2})},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			data := bytes.Repeat([]byte{0, 1}, 500)
			kit.Seek(t, tc.buf, 1, io.SeekStart)

			// --- When ---
			n, err := tc.buf.Write(data)

			// --- Then ---
			assert.NoError(t, err)
			assert.Exactly(t, 1000, n)
			assert.Exactly(t, int64(1001), kit.CurrOffset(t, tc.buf))
			want := append([]byte{0}, data...)
			assert.Exactly(t, want, kit.ReadAllFromStart(t, tc.buf))
			assert.NoError(t, tc.buf.Close())
		})
	}
}

func Test_File_Write_overrideTail(t *testing.T) {
	tt := []struct {
		testN string

		buf filer
	}{
		{"fil", TempFile(t, os.O_RDWR, []byte{0, 1, 2})},
		{"buf", With([]byte{0, 1, 2})},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			kit.Seek(t, tc.buf, 1, io.SeekStart)
			n, err := tc.buf.Write([]byte{3, 4})

			// --- Then ---
			assert.NoError(t, err)
			assert.Exactly(t, 2, n)
			assert.Exactly(t, int64(3), kit.CurrOffset(t, tc.buf))
			assert.Exactly(t, []byte{0, 3, 4}, kit.ReadAllFromStart(t, tc.buf))
			assert.NoError(t, tc.buf.Close())
		})
	}
}

func Test_File_Write_overrideMiddle(t *testing.T) {
	tt := []struct {
		testN string

		buf filer
	}{
		{"fil", TempFile(t, os.O_RDWR, []byte{0, 1, 2, 3})},
		{"buf", With([]byte{0, 1, 2, 3})},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			kit.Seek(t, tc.buf, 1, io.SeekStart)
			n, err := tc.buf.Write([]byte{4, 5})

			// --- Then ---
			assert.NoError(t, err)
			assert.Exactly(t, 2, n)
			assert.Exactly(t, int64(3), kit.CurrOffset(t, tc.buf))

			exp := []byte{0, 4, 5, 3}
			assert.Exactly(t, exp, kit.ReadAllFromStart(t, tc.buf))
			assert.NoError(t, tc.buf.Close())
		})
	}
}

func Test_File_WriteAt_zeroValue(t *testing.T) {
	tt := []struct {
		testN string

		buf filer
	}{
		{"fil", kit.TempFile(t, t.TempDir(), "")},
		{"buf", &Buffer{}},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			n, err := tc.buf.WriteAt([]byte{0, 1, 2}, 0)

			// --- Then ---
			assert.NoError(t, err)
			assert.Exactly(t, 3, n)
			assert.Exactly(t, int64(0), kit.CurrOffset(t, tc.buf))
			want := []byte{0, 1, 2}
			assert.Exactly(t, want, kit.ReadAllFromStart(t, tc.buf))
			assert.NoError(t, tc.buf.Close())
		})
	}
}

func Test_File_WriteAt_beyondCap(t *testing.T) {
	tt := []struct {
		testN string

		buf filer
	}{
		{"fil", TempFile(t, os.O_RDWR, []byte{0, 1, 2})},
		{"buf", With([]byte{0, 1, 2})},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			n, err := tc.buf.WriteAt([]byte{3, 4, 5}, 1000)

			// --- Then ---
			assert.NoError(t, err)
			assert.Exactly(t, 3, n)
			assert.Exactly(t, int64(0), kit.CurrOffset(t, tc.buf))
			want := append([]byte{0, 1, 2}, bytes.Repeat([]byte{0}, 997)...)
			want = append(want, []byte{3, 4, 5}...)
			assert.Exactly(t, want, kit.ReadAllFromStart(t, tc.buf))
			assert.NoError(t, tc.buf.Close())
		})
	}
}

func Test_File_WriteAt_append(t *testing.T) {
	tt := []struct {
		testN string

		buf filer
	}{
		{"fil", TempFile(t, os.O_RDWR, []byte{0, 1, 2})},
		{"buf", With([]byte{0, 1, 2})},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			n, err := tc.buf.WriteAt([]byte{3, 4, 5}, 3)

			// --- Then ---
			assert.NoError(t, err)
			assert.Exactly(t, 3, n)
			assert.Exactly(t, int64(0), kit.CurrOffset(t, tc.buf))

			exp := []byte{0, 1, 2, 3, 4, 5}
			assert.Exactly(t, exp, kit.ReadAllFromStart(t, tc.buf))
			assert.NoError(t, tc.buf.Close())
		})
	}
}

func Test_File_WriteAt_overrideAndExtend(t *testing.T) {
	tt := []struct {
		testN string

		buf filer
	}{
		{"fil", TempFile(t, os.O_RDWR, []byte{0, 1, 2})},
		{"buf", With([]byte{0, 1, 2})},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			data := bytes.Repeat([]byte{0, 1}, 500)
			n, err := tc.buf.WriteAt(data, 1)

			// --- Then ---
			assert.NoError(t, err)
			assert.Exactly(t, 1000, n)
			assert.Exactly(t, int64(0), kit.CurrOffset(t, tc.buf))
			want := append([]byte{0}, bytes.Repeat([]byte{0, 1}, 500)...)
			assert.Exactly(t, want, kit.ReadAllFromStart(t, tc.buf))
			assert.NoError(t, tc.buf.Close())
		})
	}
}

func Test_File_WriteAt_overrideTail(t *testing.T) {
	tt := []struct {
		testN string

		buf filer
	}{
		{"fil", TempFile(t, os.O_RDWR, []byte{0, 1, 2})},
		{"buf", With([]byte{0, 1, 2})},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			kit.Seek(t, tc.buf, 2, io.SeekStart)
			n, err := tc.buf.WriteAt([]byte{3, 4}, 1)

			// --- Then ---
			assert.NoError(t, err)
			assert.Exactly(t, 2, n)
			assert.Exactly(t, int64(2), kit.CurrOffset(t, tc.buf))
			assert.Exactly(t, []byte{0, 3, 4}, kit.ReadAllFromStart(t, tc.buf))
			assert.NoError(t, tc.buf.Close())
		})
	}
}

func Test_File_WriteAt_overrideMiddle(t *testing.T) {
	tt := []struct {
		testN string

		buf filer
	}{
		{"fil", TempFile(t, os.O_RDWR, []byte{0, 1, 2, 3})},
		{"buf", With([]byte{0, 1, 2, 3})},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			kit.Seek(t, tc.buf, 2, io.SeekStart)
			n, err := tc.buf.WriteAt([]byte{4, 5}, 1)

			// --- Then ---
			assert.NoError(t, err)
			assert.Exactly(t, 2, n)
			assert.Exactly(t, int64(2), kit.CurrOffset(t, tc.buf))

			exp := []byte{0, 4, 5, 3}
			assert.Exactly(t, exp, kit.ReadAllFromStart(t, tc.buf))
			assert.NoError(t, tc.buf.Close())
		})
	}
}

func Test_File_WriteAt_writeAtOffsetBeyondCap(t *testing.T) {
	tt := []struct {
		testN string

		buf filer
	}{
		{"fil", TempFile(t, os.O_RDWR, make([]byte, 3, 6))},
		{"buf", With(make([]byte, 3, 6))},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			kit.Seek(t, tc.buf, 0, io.SeekStart)
			n, err := tc.buf.WriteAt([]byte{1, 2}, 8)

			// --- Then ---
			assert.NoError(t, err)
			assert.Exactly(t, 2, n)
			assert.Exactly(t, int64(0), kit.CurrOffset(t, tc.buf))

			exp := []byte{0, 0, 0, 0, 0, 0, 0, 0, 1, 2}
			assert.Exactly(t, exp, kit.ReadAllFromStart(t, tc.buf))
			assert.NoError(t, tc.buf.Close())
		})
	}
}

func Test_File_WriteAt_writeAtOffsetBeyondCapOffsetCloseToLen(t *testing.T) {
	tt := []struct {
		testN string

		buf filer
	}{
		{"fil", TempFile(t, os.O_RDWR, make([]byte, 5, 7))},
		{"buf", With(make([]byte, 5, 7))},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			kit.Seek(t, tc.buf, 4, io.SeekStart)
			n, err := tc.buf.WriteAt([]byte{1, 2}, 8)

			// --- Then ---
			assert.NoError(t, err)
			assert.Exactly(t, 2, n)
			assert.Exactly(t, int64(4), kit.CurrOffset(t, tc.buf))

			exp := []byte{0, 0, 0, 0, 0, 0, 0, 0, 1, 2}
			assert.Exactly(t, exp, kit.ReadAllFromStart(t, tc.buf))
			assert.NoError(t, tc.buf.Close())
		})
	}
}

func Test_File_WriteString(t *testing.T) {
	tt := []struct {
		testN string

		buf filer
	}{
		{"fil", TempFile(t, os.O_RDWR, []byte{0, 1, 2})},
		{"buf", With([]byte{0, 1, 2})},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			kit.Seek(t, tc.buf, 1, io.SeekStart)

			// --- When ---
			n, err := tc.buf.WriteString("abc")

			// --- Then ---
			assert.NoError(t, err)
			assert.Exactly(t, 3, n)

			exp := []byte{0, 0x61, 0x62, 0x63}
			assert.Exactly(t, exp, kit.ReadAllFromStart(t, tc.buf))
			assert.Exactly(t, int64(4), kit.CurrOffset(t, tc.buf))
		})
	}
}

func Test_File_Read_zeroValue(t *testing.T) {
	tt := []struct {
		testN string

		buf filer
	}{
		{"fil", kit.TempFile(t, t.TempDir(), "")},
		{"buf", &Buffer{}},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			dst := make([]byte, 3)
			n, err := tc.buf.Read(dst)

			// --- Then ---
			assert.ErrorIs(t, err, io.EOF)
			assert.Exactly(t, 0, n)
			assert.Exactly(t, int64(0), kit.CurrOffset(t, tc.buf))
			want := []byte{0, 0, 0}
			assert.Exactly(t, want, dst)
			assert.NoError(t, tc.buf.Close())
		})
	}
}

func Test_File_Read_withSmallBuffer(t *testing.T) {
	tt := []struct {
		testN string

		buf filer
	}{
		{"fil", TempFile(t, os.O_RDONLY, []byte{0, 1, 2, 3, 4})},
		{"buf", With([]byte{0, 1, 2, 3, 4})},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			dst := make([]byte, 3)

			// --- Then ---

			// First read.
			n, err := tc.buf.Read(dst)

			assert.NoError(t, err)
			assert.Exactly(t, 3, n)
			assert.Exactly(t, int64(3), kit.CurrOffset(t, tc.buf))
			want := []byte{0, 1, 2}
			assert.Exactly(t, want, dst)

			// Second read.
			n, err = tc.buf.Read(dst)

			assert.NoError(t, err)
			assert.Exactly(t, 2, n)
			assert.Exactly(t, int64(5), kit.CurrOffset(t, tc.buf))
			want = []byte{3, 4, 2}
			assert.Exactly(t, want, dst)

			// Third read.
			n, err = tc.buf.Read(dst)

			assert.ErrorIs(t, err, io.EOF)
			assert.Exactly(t, 0, n)
			assert.Exactly(t, int64(5), kit.CurrOffset(t, tc.buf))
			want = []byte{3, 4, 2}
			assert.Exactly(t, want, dst)

			assert.NoError(t, tc.buf.Close())
		})
	}
}

func Test_File_Read_beyondLen(t *testing.T) {
	tt := []struct {
		testN string

		buf filer
	}{
		{"fil", TempFile(t, os.O_RDWR, []byte{0, 1, 2})},
		{"buf", With([]byte{0, 1, 2})},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			kit.Seek(t, tc.buf, 5, io.SeekStart)

			// --- When ---
			dst := make([]byte, 3)
			n, err := tc.buf.Read(dst)

			// --- Then ---
			assert.ErrorIs(t, err, io.EOF)
			assert.Exactly(t, 0, n)
			assert.Exactly(t, int64(5), kit.CurrOffset(t, tc.buf))
			want := []byte{0, 0, 0}
			assert.Exactly(t, want, dst)
			assert.NoError(t, tc.buf.Close())
		})
	}
}

func Test_File_Read_bigBuffer(t *testing.T) {
	tt := []struct {
		testN string

		buf filer
	}{
		{"fil", TempFile(t, os.O_RDWR, []byte{0, 1, 2})},
		{"buf", With([]byte{0, 1, 2})},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			dst := make([]byte, 6)
			n, err := tc.buf.Read(dst)

			// --- Then ---
			assert.NoError(t, err)
			assert.Exactly(t, 3, n)
			assert.Exactly(t, int64(3), kit.CurrOffset(t, tc.buf))
			want := []byte{0, 1, 2, 0, 0, 0}
			assert.Exactly(t, want, dst)
			assert.NoError(t, tc.buf.Close())
		})
	}
}

func Test_File_Read_readAll(t *testing.T) {
	tt := []struct {
		testN string

		buf filer
	}{
		{"fil", TempFile(t, os.O_RDWR, []byte{0, 1, 2})},
		{"buf", With([]byte{0, 1, 2})},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			dst := make([]byte, 3, 4)

			// --- When ---
			n, err := tc.buf.Read(dst)

			// --- Then ---
			assert.NoError(t, err)
			assert.Exactly(t, 3, n)
			assert.Exactly(t, int64(3), kit.CurrOffset(t, tc.buf))
			assert.Exactly(t, []byte{0, 1, 2}, dst)
			assert.NoError(t, tc.buf.Close())
		})
	}
}

func Test_File_Read_readHead(t *testing.T) {
	tt := []struct {
		testN string

		buf filer
	}{
		{"fil", TempFile(t, os.O_RDWR, []byte{0, 1, 2, 4})},
		{"buf", With([]byte{0, 1, 2, 4})},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			dst := make([]byte, 3, 4)

			// --- When ---
			n, err := tc.buf.Read(dst)

			// --- Then ---
			assert.NoError(t, err)
			assert.Exactly(t, 3, n)
			assert.Exactly(t, int64(3), kit.CurrOffset(t, tc.buf))
			assert.Exactly(t, []byte{0, 1, 2}, dst)
			assert.NoError(t, tc.buf.Close())
		})
	}
}

func Test_File_Read_readTail(t *testing.T) {
	tt := []struct {
		testN string

		buf filer
	}{
		{"fil", TempFile(t, os.O_RDWR, []byte{0, 1, 2})},
		{"buf", With([]byte{0, 1, 2})},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			kit.Seek(t, tc.buf, 1, io.SeekStart)
			dst := make([]byte, 2, 3)

			// --- When ---
			n, err := tc.buf.Read(dst)

			// --- Then ---
			assert.NoError(t, err)
			assert.Exactly(t, 2, n)
			assert.Exactly(t, int64(3), kit.CurrOffset(t, tc.buf))
			assert.Exactly(t, []byte{1, 2}, dst)
			assert.NoError(t, tc.buf.Close())
		})
	}
}

func Test_File_ReadAt_beyondLen(t *testing.T) {
	tt := []struct {
		testN string

		buf filer
	}{
		{"fil", TempFile(t, os.O_RDWR, []byte{0, 1, 2})},
		{"buf", With([]byte{0, 1, 2})},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			dst := make([]byte, 4)
			n, err := tc.buf.ReadAt(dst, 6)

			// --- Then ---
			assert.ErrorIs(t, err, io.EOF)
			assert.Exactly(t, 0, n)
			assert.Exactly(t, int64(0), kit.CurrOffset(t, tc.buf))
			want := []byte{0, 0, 0, 0}
			assert.Exactly(t, want, dst)
			assert.NoError(t, tc.buf.Close())
		})
	}
}

func Test_File_ReadAt_bigBuffer(t *testing.T) {
	tt := []struct {
		testN string

		buf filer
	}{
		{"fil", TempFile(t, os.O_RDWR, []byte{0, 1, 2})},
		{"buf", With([]byte{0, 1, 2})},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			kit.Seek(t, tc.buf, 1, io.SeekStart)
			dst := make([]byte, 4)

			// --- When ---
			n, err := tc.buf.ReadAt(dst, 0)

			// --- Then ---
			assert.ErrorIs(t, err, io.EOF)
			assert.Exactly(t, 3, n)
			assert.Exactly(t, int64(1), kit.CurrOffset(t, tc.buf))
			want := []byte{0, 1, 2, 0}
			assert.Exactly(t, want, dst)
			assert.NoError(t, tc.buf.Close())
		})
	}
}

func Test_File_ReadAt_readAll(t *testing.T) {
	tt := []struct {
		testN string

		buf filer
	}{
		{"fil", TempFile(t, os.O_RDWR, []byte{0, 1, 2})},
		{"buf", With([]byte{0, 1, 2})},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			kit.Seek(t, tc.buf, 1, io.SeekStart)
			dst := []byte{0, 1, 2}

			// --- When ---
			n, err := tc.buf.ReadAt(dst, 0)

			// --- Then ---
			assert.NoError(t, err)
			assert.Exactly(t, 3, n)
			assert.Exactly(t, int64(1), kit.CurrOffset(t, tc.buf))
			assert.Exactly(t, []byte{0, 1, 2}, dst)
			assert.NoError(t, tc.buf.Close())
		})
	}
}

func Test_File_ReadAt_readHead(t *testing.T) {
	tt := []struct {
		testN string

		buf filer
	}{
		{"fil", TempFile(t, os.O_RDWR, []byte{0, 1, 2})},
		{"buf", With([]byte{0, 1, 2})},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			kit.Seek(t, tc.buf, 1, io.SeekStart)
			dst := make([]byte, 2, 3)

			// --- When ---
			n, err := tc.buf.ReadAt(dst, 0)

			// --- Then ---
			assert.NoError(t, err)
			assert.Exactly(t, 2, n)
			assert.Exactly(t, int64(1), kit.CurrOffset(t, tc.buf))
			assert.Exactly(t, []byte{0, 1}, dst)
			assert.NoError(t, tc.buf.Close())
		})
	}
}

func Test_File_ReadAt_readTail(t *testing.T) {
	tt := []struct {
		testN string

		buf filer
	}{
		{"fil", TempFile(t, os.O_RDWR, []byte{0, 1, 2})},
		{"buf", With([]byte{0, 1, 2})},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			kit.Seek(t, tc.buf, 2, io.SeekStart)
			dst := make([]byte, 2, 3)

			// --- When ---
			n, err := tc.buf.ReadAt(dst, 1)

			// --- Then ---
			assert.NoError(t, err)
			assert.Exactly(t, 2, n)
			assert.Exactly(t, int64(2), kit.CurrOffset(t, tc.buf))
			assert.Exactly(t, []byte{1, 2}, dst)
			assert.NoError(t, tc.buf.Close())
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
			fil := TempFile(t, os.O_RDWR, []byte{0, 1, 2, 3})
			kit.Seek(t, fil, 1, io.SeekStart)

			buf := With([]byte{0, 1, 2, 3})
			kit.Seek(t, buf, 1, io.SeekStart)

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

func Test_File_Seek_beyondLen(t *testing.T) {
	tt := []struct {
		testN string

		buf filer
	}{
		{"fil", TempFile(t, os.O_RDWR, []byte{0, 1, 2})},
		{"buf", With([]byte{0, 1, 2})},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			n, err := tc.buf.Seek(5, io.SeekStart)

			// --- Then ---
			assert.NoError(t, err)
			assert.Exactly(t, int64(5), n)
		})
	}
}

func Test_File_Truncate_toZero(t *testing.T) {
	tt := []struct {
		testN string

		buf filer
	}{
		{"fil", TempFile(t, os.O_RDWR, []byte{0, 1, 2, 3})},
		{"buf", With([]byte{0, 1, 2, 3})},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			err := tc.buf.Truncate(0)

			// --- Then ---
			assert.NoError(t, err)
			assert.Exactly(t, int64(0), kit.CurrOffset(t, tc.buf))
			assert.Exactly(t, []byte{}, kit.ReadAllFromStart(t, tc.buf))
			assert.NoError(t, tc.buf.Close())
		})
	}
}

func Test_File_Truncate_toOne(t *testing.T) {
	tt := []struct {
		testN string

		buf filer
	}{
		{"fil", TempFile(t, os.O_RDWR, []byte{0, 1, 2, 3})},
		{"buf", With([]byte{0, 1, 2, 3})},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			err := tc.buf.Truncate(1)

			// --- Then ---
			assert.NoError(t, err)
			assert.Exactly(t, int64(0), kit.CurrOffset(t, tc.buf))
			assert.Exactly(t, []byte{0}, kit.ReadAllFromStart(t, tc.buf))
			assert.NoError(t, tc.buf.Close())
		})
	}
}

func Test_File_Truncate_toZeroAndWrite(t *testing.T) {
	tt := []struct {
		testN string

		buf filer
	}{
		{"fil", TempFile(t, os.O_RDWR, []byte{0, 1, 2, 3})},
		{"buf", With([]byte{0, 1, 2, 3})},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			err := tc.buf.Truncate(0)
			assert.NoError(t, err)

			n, err := tc.buf.Write([]byte{4, 5})

			// --- Then ---
			assert.NoError(t, err)
			assert.Exactly(t, 2, n)
			assert.Exactly(t, int64(2), kit.CurrOffset(t, tc.buf))
			assert.Exactly(t, []byte{4, 5}, kit.ReadAllFromStart(t, tc.buf))
			assert.NoError(t, tc.buf.Close())
		})
	}
}

func Test_File_Truncate_beyondLenAndWrite(t *testing.T) {
	tt := []struct {
		testN string

		buf filer
	}{
		{"fil", TempFile(t, os.O_RDWR|os.O_APPEND, []byte{0, 1, 2, 3})},
		{"buf", With([]byte{0, 1, 2, 3}, Append)},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			kit.Seek(t, tc.buf, 1, io.SeekStart)

			// --- When ---
			assert.NoError(t, tc.buf.Truncate(8))
			n, err := tc.buf.Write([]byte{4, 5})

			// --- Then ---
			assert.NoError(t, err)
			assert.Exactly(t, 2, n)
			assert.Exactly(t, int64(10), kit.CurrOffset(t, tc.buf))

			exp := []byte{0, 1, 2, 3, 0, 0, 0, 0, 4, 5}
			assert.Exactly(t, exp, kit.ReadAllFromStart(t, tc.buf))
			assert.NoError(t, tc.buf.Close())
		})
	}
}

func Test_File_Truncate_beyondCapAndWrite(t *testing.T) {
	// Initial buffer value.
	in := make([]byte, 4, 8)
	in[0] = 0
	in[1] = 1
	in[2] = 2
	in[3] = 3

	tt := []struct {
		testN string

		buf filer
	}{
		{"fil", TempFile(t, os.O_RDWR|os.O_APPEND, in)},
		{"buf", With(in, Append)},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			assert.NoError(t, tc.buf.Truncate(10))
			n, err := tc.buf.Write([]byte{4, 5})

			// --- Then ---
			assert.NoError(t, err)
			assert.Exactly(t, 2, n)
			assert.Exactly(t, int64(12), kit.CurrOffset(t, tc.buf))
			want := []byte{0, 1, 2, 3, 0, 0, 0, 0, 0, 0, 4, 5}
			assert.Exactly(t, want, kit.ReadAllFromStart(t, tc.buf))
			assert.NoError(t, tc.buf.Close())
		})
	}
}

func Test_File_Truncate_extendBeyondLenResetAndWrite(t *testing.T) {
	tt := []struct {
		testN string

		buf filer
	}{
		{"fil", TempFile(t, os.O_RDWR|os.O_APPEND, []byte{0, 1, 2, 3})},
		{"buf", With([]byte{0, 1, 2, 3}, Append)},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			assert.NoError(t, tc.buf.Truncate(8))
			assert.NoError(t, tc.buf.Truncate(0))
			n, err := tc.buf.Write([]byte{4, 5})

			// --- Then ---
			assert.NoError(t, err)
			assert.Exactly(t, 2, n)
			assert.Exactly(t, int64(2), kit.CurrOffset(t, tc.buf))
			assert.Exactly(t, []byte{4, 5}, kit.ReadAllFromStart(t, tc.buf))
			assert.NoError(t, tc.buf.Close())
		})
	}
}

func Test_File_Truncate_edgeCaseWhenSizeEqualsLength(t *testing.T) {
	tt := []struct {
		testN string

		buf filer
	}{
		{"fil", TempFile(t, os.O_RDWR|os.O_APPEND, []byte{0, 1, 2, 3})},
		{"buf", With([]byte{0, 1, 2, 3}, Append)},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			assert.NoError(t, tc.buf.Truncate(4))
			n, err := tc.buf.Write([]byte{4, 5})

			// --- Then ---
			assert.NoError(t, err)
			assert.Exactly(t, 2, n)
			assert.Exactly(t, int64(6), kit.CurrOffset(t, tc.buf))

			exp := []byte{0, 1, 2, 3, 4, 5}
			assert.Exactly(t, exp, kit.ReadAllFromStart(t, tc.buf))
			assert.NoError(t, tc.buf.Close())
		})
	}
}
