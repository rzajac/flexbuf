# Flexible bytes buffer.

[![Go Report Card](https://goreportcard.com/badge/github.com/rzajac/flexbuf)](https://goreportcard.com/report/github.com/rzajac/flexbuf)
[![GoDoc](https://img.shields.io/badge/api-Godoc-blue.svg)](https://pkg.go.dev/github.com/rzajac/flexbuf)

Package `flexbuf` provides bytes buffer implementing many data access and 
manipulation interfaces.

    io.Writer
    io.WriterAt
    io.Reader
    io.ReaderAt
    io.ReaderFrom
    io.Seeker
    io.Closer
    
Additionally, `flexbuf` provides `Truncate(size int64) error` method to make 
it almost a drop in replacement for `os.File`.

The `flexbuf.Buffer` also uses `sync.Pool` so when you create a lot of buffers 
it will reuse buffers from the pool - see [flexbuf.New](https://pkg.go.dev/github.com/rzajac/flexbuf#New)
constructor.

# Installation.

```
go get github.com/rzajac/zrr
```

# Examples

```
buf := &flexbuf.Buffer{}

_, _ = buf.Write([]byte{0, 1, 2, 3})
_, _ = buf.Seek(-2, io.SeekEnd)
_, _ = buf.Write([]byte{4, 5})
_, _ = buf.Seek(0, io.SeekStart)

data, _ := ioutil.ReadAll(buf)
fmt.Println(data)

// Output: [0 1 4 5]
```

# How is it different from `bytes.Buffer`?

The `bytes.Buffer` always reads from current offset and writes to the end of 
the buffer, `flexbuf` behaves more like a file it reads and writes at current 
offset. Also `bytes.Buffer` doesn't implement interfaces:

- `io.WriterAt`
- `io.ReaderAt`
- `io.Seeker`
- `io.Closer`

or methods:

- `Truncate`

# Can I use `flexbuf.Buffer` as a replacement for `os.File`?

It depends. Even though `flexbuf.Buffer` probably implements all the methods 
you need to use it as a replacement for `os.File` there are some minor 
differences:

- `Truncate` method does not return `os.PathError` instances.
- `WriteAt` will not return error when used on an instance created with
    `flexbuf.New(flexbuf.Append)` or `flexbuf.With(myBuf, flexbuf.Append)`.

# License

BSD-2-Clause