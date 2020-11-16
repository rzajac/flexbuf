// Package flexbuf provides bytes buffer implementing many data access and
// manipulation interfaces.
//
//    io.Writer
//    io.WriterAt
//    io.Reader
//    io.ReaderAt
//    io.ReaderFrom
//    io.Seeker
//
// Additionally, `flexbuf` provides `Truncate(size int64) error` method to make
// it almost a drop in replacement for `os.File`.
//

package flexbuf

import (
	"bytes"
	"errors"
	"io"
	"os"
	"sync"
)

// ErrOutOfBounds is returned for invalid offsets.
var ErrOutOfBounds = errors.New("offset out of bounds")

// pool of byte buffers.
var pool = sync.Pool{
	New: func() interface{} {
		return make([]byte, bytes.MinRead)
	},
}

// Offset is the constructor option setting the initial buffer offset to off.
func Offset(off int) func(*Buffer) error {
	return func(b *Buffer) error {
		if off < 0 || off > len(b.buf) {
			return ErrOutOfBounds
		}
		b.off = off
		return nil
	}
}

// Append is the constructor option setting the initial offset
// to the end of the buffer. Append should be the last option on the
// option list.
func Append(buf *Buffer) error {
	buf.off = len(buf.buf)
	return nil
}

// A Buffer is a variable-sized buffer of bytes.
// The zero value for Buffer is an empty buffer ready to use.
type Buffer struct {
	// Set to false when underlying buffer was allocated from the pool.
	external bool
	// Current offset for read and write operations.
	off int
	// Underlying buffer.
	buf []byte
}

// New returns new instance of the Buffer. The difference between New and
// using zero value buffer is that New will get the initial buffer from
// the pool.
func New(opts ...func(buffer *Buffer) error) (*Buffer, error) {
	buf := pool.Get().([]byte)[:0]
	b, err := With(buf, opts...)
	if err != nil {
		return nil, err
	}
	b.external = false
	return b, err
}

// With creates new instance of Buffer initialized with data.
func With(data []byte, opts ...func(*Buffer) error) (*Buffer, error) {
	b := &Buffer{
		external: true,
		buf:      data,
	}

	for _, opt := range opts {
		if err := opt(b); err != nil {
			return nil, err
		}
	}

	return b, nil
}

// Write writes the contents of p to the buffer at current offset, growing
// the buffer as needed. The return value n is the length of p; err is
// always nil.
func (b *Buffer) Write(p []byte) (int, error) {
	return b.write(p), nil
}

// write writes p at offset b.off.
func (b *Buffer) write(p []byte) int {
	b.grow(len(p))
	n := copy(b.buf[b.off:], p)
	b.off += n
	return n
}

// WriteAt writes len(p) bytes to the buffer starting at byte offset off.
// It returns the number of bytes written; err is always nil. It does not
// change the offset.
func (b *Buffer) WriteAt(p []byte, off int64) (int, error) {
	prev := b.off
	b.off = int(off)
	n := b.write(p)
	b.off = prev
	return n, nil
}

// Read reads the next len(p) bytes from the buffer or until the buffer
// is drained. The return value is the number of bytes read. If the
// buffer has no data to return, err is io.EOF (unless len(p) is zero);
// otherwise it is nil.
func (b *Buffer) Read(p []byte) (int, error) {
	// Nothing more to read.
	if len(p) > 0 && b.off >= len(b.buf) {
		return 0, io.EOF
	}
	n := copy(p, b.buf[b.off:])
	b.off += n
	if len(b.buf[b.off:]) > 0 {
		return n, nil
	}
	return n, io.EOF
}

// ReadAt reads len(p) bytes from the buffer starting at byte offset off.
// It returns the number of bytes read and the error, if any.
// ReadAt always returns a non-nil error when n < len(p). It does not
// change the offset.
func (b *Buffer) ReadAt(p []byte, off int64) (int, error) {
	if off >= int64(len(b.buf)) {
		return 0, ErrOutOfBounds
	}
	prev := b.off
	defer func() { b.off = prev }()
	b.off = int(off)
	n, err := b.Read(p)
	if err != nil {
		return n, err
	}
	return n, nil
}

// ReadFrom reads data from r until EOF and appends it to the buffer at b.off,
// growing the buffer as needed. The return value is the number of bytes read.
// Any error except io.EOF encountered during the read is also returned. If the
// buffer becomes too large, ReadFrom will panic with ErrTooLarge.
func (b *Buffer) ReadFrom(r io.Reader) (int64, error) {
	var total int
	for {
		// Length before growing the buffer.
		l := len(b.buf)

		// Make sure we can fit MinRead between b.off and new buffer length.
		b.grow(bytes.MinRead)

		// Because io.Read documentation says: "Even if Read returns
		// n < len(p), it may use all of p as scratch space during the call."
		// we can't pass our buffer to read because it might change parts of it
		// not involved in read operation. We will use temporary bytes buffer
		// from the pool for reading and then copy read bytes to actual buffer.
		tmp := pool.Get().([]byte)

		n, err := r.Read(tmp)
		copy(b.buf[b.off:], tmp[:n])
		zeroOutSlice(tmp[:n])
		pool.Put(tmp)

		b.off += n
		total += n

		// In case we have read less them MinRead bytes
		// we have to set proper buffer length.
		b.buf = b.buf[:l+n]

		// The io.EOF is not an error.
		if err == io.EOF {
			return int64(total), nil
		}
		if err != nil {
			return int64(total), err
		}
	}
}

// Seek sets the offset for the next Read or Write on the buffer to offset,
// interpreted according to whence: 0 means relative to the origin of the file,
// 1 means relative to the current offset, and 2 means relative to the end.
// It returns the new offset and an error, if any.
func (b *Buffer) Seek(offset int64, whence int) (int64, error) {
	var off int
	switch whence {
	case io.SeekStart:
		off = int(offset)
	case io.SeekCurrent:
		off = b.off + int(offset)
	case io.SeekEnd:
		off = len(b.buf) + int(offset)
	}

	if off < 0 {
		return 0, os.ErrInvalid
	}
	b.off = off

	return int64(b.off), nil
}

// Truncate changes the size of the buffer discarding bytes at offsets greater
// then size. It does not change the offset.
func (b *Buffer) Truncate(size int64) error {
	if size < 0 {
		return os.ErrInvalid
	}

	// Extend the size of the buffer.
	if int(size) > len(b.buf) {
		b.grow(int(size) - len(b.buf))
		return nil
	}

	// Reduce the size of the buffer.
	zeroOutSlice(b.buf[size:])
	b.buf = b.buf[:size]

	return nil
}

// tryGrowByReslice is a inlineable version of grow for the fast-case where the
// internal buffer only needs to be resliced. It returns whether it succeeded.
func (b *Buffer) tryGrowByReslice(n int) bool {
	// No need to do anything if there is enough space
	// between current offset and the length of the buffer.
	if n <= len(b.buf)-b.off {
		return true
	}

	if n <= cap(b.buf)-b.off {
		b.buf = b.buf[:b.off+n]
		return true
	}
	return false
}

// grow grows the buffer capacity to guarantee space for n more bytes. In
// another words it makes sure there is n bytes between b.off and buffer
// capacity. It's worth noting that after calling this method the len(b.buf)
// changes. If the buffer can't grow it will panic with ErrTooLarge.
func (b *Buffer) grow(n int) {
	// Try to grow by means of a reslice.
	if ok := b.tryGrowByReslice(n); ok {
		return
	}

	// The total capacity y of the buffer.
	c := cap(b.buf)
	// The real capacity of the buffer.
	// We keep all the bytes before b.off when writing new bytes.
	rc := c - b.off
	// How much do we have to extend capacity to
	// accommodate n additional bytes.
	ex := c + n - rc

	// Allocate buffer which is big enough for what we have
	// in the buffer [0:b.off] and n additional bytes.
	tmp := makeSlice(ex)
	copy(tmp, b.buf)
	b.buf = tmp
}

// makeSlice allocates a slice of size n. If the allocation fails, it panics
// with ErrTooLarge.
func makeSlice(n int) []byte {
	// If the make fails, give a known error.
	defer func() {
		if recover() != nil {
			panic(bytes.ErrTooLarge)
		}
	}()
	return make([]byte, n)
}

// Offset returns the current offset.
func (b *Buffer) Offset() int {
	return b.off
}

// Len returns the number of bytes in the buffer.
func (b *Buffer) Len() int {
	return len(b.buf)
}

// Cap returns the capacity of the buffer's underlying byte slice, that is, the
// total space allocated for the buffer's data.
func (b *Buffer) Cap() int {
	return cap(b.buf)
}

// Close sets offset to zero, if underlying buffer was allocated from the
// pool it is zeroed out and put back to the pool. It always returns nil error.
func (b *Buffer) Close() error {
	b.off = 0
	if !b.external {
		zeroOutSlice(b.buf)
		pool.Put(b.buf)
	} else {
		b.buf = nil
	}
	return nil
}
