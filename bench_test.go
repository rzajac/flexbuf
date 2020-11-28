package flexbuf_test

import (
	"bytes"
	"testing"

	"github.com/rzajac/flexbuf"
)

//goland:noinspection GoUnusedGlobalVariable
var bufferWrite int

func BenchmarkWrite(b *testing.B) {
	data := make([]byte, 1<<15)

	b.Run("flexbuf", func(b *testing.B) {
		b.ReportAllocs()
		b.StopTimer()
		var n int

		b.StartTimer()
		for i := 0; i < b.N; i++ {
			buf := &flexbuf.Buffer{}
			n, _ = buf.Write(data)
		}
		bufferWrite = n
	})

	b.Run("bytes", func(b *testing.B) {
		b.ReportAllocs()
		b.StopTimer()
		var n int

		b.StartTimer()
		for i := 0; i < b.N; i++ {
			buf := &bytes.Buffer{}
			n, _ = buf.Write(data)
		}
		bufferWrite = n
	})
}

//goland:noinspection GoUnusedGlobalVariable
var bufferWriteByte error

func BenchmarkWriteByte(b *testing.B) {
	b.Run("flexbuf", func(b *testing.B) {
		b.ReportAllocs()
		b.StopTimer()
		var err error

		b.StartTimer()
		for i := 0; i < b.N; i++ {
			buf := &flexbuf.Buffer{}
			err = buf.WriteByte(1)
		}
		bufferWriteByte = err
	})

	b.Run("bytes", func(b *testing.B) {
		b.ReportAllocs()
		b.StopTimer()
		var err error

		b.StartTimer()
		for i := 0; i < b.N; i++ {
			buf := &bytes.Buffer{}
			err = buf.WriteByte(1)
		}
		bufferWriteByte = err
	})
}

//goland:noinspection GoUnusedGlobalVariable
var bufferWriteString int

func BenchmarkWriteString(b *testing.B) {
	b.Run("flexbuf", func(b *testing.B) {
		b.ReportAllocs()
		b.StopTimer()
		var n int

		b.StartTimer()
		for i := 0; i < b.N; i++ {
			buf := &flexbuf.Buffer{}
			n, _ = buf.WriteString("abcdefghijkl")
		}
		bufferWriteString = n
	})

	b.Run("bytes", func(b *testing.B) {
		b.ReportAllocs()
		b.StopTimer()
		var n int

		b.StartTimer()
		for i := 0; i < b.N; i++ {
			buf := &bytes.Buffer{}
			n, _ = buf.WriteString("abcdefghijkl")
		}
		bufferWriteString = n
	})
}

//goland:noinspection GoUnusedGlobalVariable
var bufferReadFrom int64

func BenchmarkReadFrom(b *testing.B) {
	data := make([]byte, 1<<15)

	b.Run("flexbuf", func(b *testing.B) {
		b.ReportAllocs()
		b.StopTimer()
		var n int64
		src := bytes.NewReader(data)

		b.StartTimer()
		for i := 0; i < b.N; i++ {
			buf := &flexbuf.Buffer{}
			n, _ = buf.ReadFrom(src)
			src.Reset(data)
		}
		bufferReadFrom = n
	})

	b.Run("bytes", func(b *testing.B) {
		b.ReportAllocs()
		b.StopTimer()
		var n int64
		src := bytes.NewReader(data)

		b.StartTimer()
		for i := 0; i < b.N; i++ {
			buf := &bytes.Buffer{}
			n, _ = buf.ReadFrom(src)
			src.Reset(data)
		}
		bufferReadFrom = n
	})
}
