package flexbuf

import (
	"os"
	"testing"

	kit "github.com/rzajac/testkit"
)

// TempFile creates and opens temporary file with contents from data slice.
//
// The flag parameter is the same as in os.Open.
//
// On error function calls t.Fatal. Function registers function in
// test cleanup to close the file and remove it.
func TempFile(t *testing.T, flag int, data []byte) *os.File {
	t.Helper()
	pth := kit.TempFileBuf(t, t.TempDir(), data)
	fil, err := os.OpenFile(pth, flag, 0666)
	if err != nil {
		t.Fatal(err)
	}
	return fil
}
