package flexbuf_test

import (
	"fmt"
	"io"
	"io/ioutil"

	"github.com/rzajac/flexbuf"
)

func ExampleBuffer() {
	buf := &flexbuf.Buffer{}

	_, _ = buf.Write([]byte{0, 1, 2, 3})
	_, _ = buf.Seek(-2, io.SeekEnd)
	_, _ = buf.Write([]byte{4, 5})
	_, _ = buf.Seek(0, io.SeekStart)

	data, _ := ioutil.ReadAll(buf)
	fmt.Println(data)

	// Output: [0 1 4 5]
}
