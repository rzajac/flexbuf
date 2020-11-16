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
    
Additionally, `flexbuf` provides `Truncate(size int64) error` method to make 
it almost a drop in replacement for `os.File`.

# Installation.

```
go get github.com/rzajac/zrr
```

# Examples

```go
buf := &flexbuf.Buffer{}
buf.Write([]byte{0, 1, 2, 3})
buf.Seek(-2, io.SeekEnd)
buf.Write([]byte{4, 5})
buf.Seek(0, io.SeekStart)

data, _ := ioutil.ReadAll(buf)
fmt.Println(data)

// Output: [0 1 4 5]
```

# License

BSD-2-Clause