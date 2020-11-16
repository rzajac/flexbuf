package flexbuf_test

import (
	"fmt"
	"io"
	"io/ioutil"

	"github.com/rzajac/flexbuf"
)

func ExampleBuffer() {
	buf := &flexbuf.Buffer{}
	buf.Write([]byte{0, 1, 2, 3})
	buf.Seek(-2, io.SeekEnd)
	buf.Write([]byte{4, 5})
	buf.Seek(0, io.SeekStart)

	data, _ := ioutil.ReadAll(buf)
	fmt.Println(data)

	// Output: [0 1 4 5]
}
