// Package goback implements buffered reader in backward order.
// Chunks are read from left to right but moving offset from right to left.
// Backward reader may be useful for reading size endian records in an append only file,
// specially when dangling links are included in records.
// A potential use case for this backward reader would be the implementation of a BTree
// over an append only file as implemented in Couchdb.
//
// package main
//
// import (
//   "os"
//	 "github.com/jeroiraz/goback"
// )
//
// func main() {
//	 f, _ := os.Open("file")
//	 defer f.Close()
//
//	 reader, _ := goback.NewFileReader(f)
//
//   // Read last 32 bytes from input file
//	 buf := make([]byte, 32)
//	 reader.Read(buf)
// }

package goback

import (
	"errors"
	"io"
	"os"

	"github.com/jeroiraz/goback/backbuf"
)

// ErrInvalidArguments, ErrNoContent, ErrInternal error list
var (
	ErrInvalidArguments = errors.New("goback: invalid arguments")
	ErrNoContent        = errors.New("goback: no content to be read")
	ErrInternal         = errors.New("goback: unexpected internal error")
)

const (
	minBufSize     = 1
	defaultBufSize = 4096
)

// Reader implements a buffered reader however read happens in backward manner
// Bytes are read from left to right but offset is updaqted in descending order
type Reader struct {
	ir  RndReader
	off int64
	buf *backbuf.Buffer
}

// RndReader specifies the required method for reading source data
type RndReader interface {
	ReadAt(b []byte, off int64) (n int, err error)
}

// NewFileReader creates a reader starting from the last byte of the file
func NewFileReader(f *os.File) (*Reader, error) {
	return NewFileReaderSize(f, defaultBufSize)
}

// NewFileReaderSize creates a backward reader with buffer size `bufSize`
func NewFileReaderSize(f *os.File, bufSize int) (*Reader, error) {
	off, err := f.Seek(0, io.SeekEnd)
	if err != nil {
		return nil, err
	}
	return NewReader(f, off, bufSize)
}

// NewReader creates a backward reader starting at offset `off` and buffer size `bufSize`
func NewReader(ir RndReader, off int64, bufSize int) (*Reader, error) {
	if bufSize < minBufSize {
		return nil, ErrInvalidArguments
	}

	buf, err := backbuf.New(bufSize)
	if err != nil {
		return nil, err
	}

	r := &Reader{
		ir:  ir,
		off: off,
		buf: buf,
	}

	return r, nil
}

// Offset returns the maximun number of bytes that can be read
func (r *Reader) Offset() int64 {
	return r.off + int64(r.buf.ReadAvailability())
}

// ReadAt reads at most `len(b)` bytes starting at offset `off`
func (r *Reader) ReadAt(b []byte, off int64) (int, error) {
	if off < r.off {
		r.buf.Reset()
	} else {
		if off > r.off+int64(r.buf.ReadAvailability()) {
			r.buf.Reset()
		} else {
			r.buf.Omit(int(off - r.off))
		}
	}
	r.off = off

	return r.Read(b)
}

// Read reads at most `len(b)` bytes starting from current offset
func (r *Reader) Read(b []byte) (int, error) {
	if b == nil || len(b) == 0 {
		return 0, ErrInvalidArguments
	}

	toRead := len(b)
	if int64(toRead) > r.Offset() {
		toRead = int(r.Offset())
	}

	read := 0

	for read != toRead {
		if r.buf.ReadAvailability() == 0 {
			maxReadSize := r.buf.Size()
			if r.off < int64(maxReadSize) {
				maxReadSize = int(r.off)
			}

			roff := r.off - int64(maxReadSize)
			bs := make([]byte, maxReadSize)
			nr, err := r.ir.ReadAt(bs, roff)

			if err != nil {
				return 0, err
			}

			nw, err := r.buf.Write(bs[:nr])
			if err != nil {
				return 0, err
			}

			if nr != nw {
				return 0, ErrInternal
			}

			r.off = roff
		}

		rs := min(toRead-read, r.buf.ReadAvailability())
		n, err := r.buf.Read(b[toRead-read-rs : toRead-read])

		if err != nil {
			return 0, err
		}

		if n != rs {
			return 0, ErrInternal
		}

		read += rs
	}

	if read < len(b) {
		return read, ErrNoContent
	}

	return read, nil
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
