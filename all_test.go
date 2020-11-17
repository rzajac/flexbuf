package flexbuf

import (
	"io"
	"io/ioutil"
	"os"
	"testing"
)

// tmpFile creates and opens temporary file and returns its descriptor.
// On error function calls t.Fatal. Function registers function in
// test cleanup to close the file and remove it.
func tmpFile(t *testing.T) *os.File {
	t.Helper()
	f, err := ioutil.TempFile("", "tmp_*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = f.Close()
		if err := os.Remove(f.Name()); err != nil {
			t.Fatal(err)
		}
	})
	return f
}

// tmpFile creates and opens temporary file with contents from data slice.
//
// The flag parameter is the same as in os.Open.
//
// After writing the data file is closed and reopened before returning the
// descriptor. On error function calls t.Fatal. Function registers function in
// test cleanup to close the file and remove it.
func tmpFileData(t *testing.T, flag int, data []byte) *os.File {
	t.Helper()
	f, err := ioutil.TempFile("", "tmp_*")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write(data); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	if f, err = os.OpenFile(f.Name(), flag, 0666); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = f.Close()
		if err := os.Remove(f.Name()); err != nil {
			t.Fatal(err)
		}
	})
	return f
}

// mustReadWhole seeks to the beginning of the rs and reads till the EOF.
func mustReadWhole(rs io.ReadSeeker) []byte {
	mustSeek(rs, 0, io.SeekStart)
	data, err := ioutil.ReadAll(rs)
	if err != nil {
		panic(err)
	}
	return data
}

// mustCurrOffset returns the current offset of the seeker. Panics on error.
func mustCurrOffset(s io.Seeker) int64 {
	return mustSeek(s, 0, io.SeekCurrent)
}

// mustSeek sets the offset for the next Read or Write to offset,
// interpreted according to whence.
// mustSeek returns the new offset relative to the start of the s.
// Panics on error.
func mustSeek(s io.Seeker, offset int64, whence int) int64 {
	off, err := s.Seek(offset, whence)
	if err != nil {
		panic(err)
	}
	return off
}

// mustFileSize returns file size. Panics on erorr.
func mustFileSize(f *os.File) int64 {
	s, err := f.Stat()
	if err != nil {
		panic(err)
	}
	return s.Size()
}
