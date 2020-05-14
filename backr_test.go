package goback

import (
	"bytes"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"
	"time"
)

type InMemReader struct {
	buf []byte
}

func (r *InMemReader) ReadAt(b []byte, off int64) (n int, err error) {
	rs := min(len(r.buf)-int(off), len(b))
	copy(b[:rs], r.buf[off:int(off)+rs])

	if len(b) < rs {
		return rs, ErrNoContent
	}

	return rs, nil
}

func randomBytes(size int) []byte {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, size)
	rand.Read(b)
	for i, v := range b {
		if v == 0 {
			b[i]++
		}
	}
	return b
}

// TestReaderEquals checks read content matches source content in the very same order
func TestReaderEquals(t *testing.T) {
	b := randomBytes(1024)
	r := &InMemReader{buf: b}

	readSize := 16
	reader, err := NewReader(r, int64(len(b)), readSize)

	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}

	if reader == nil {
		t.Errorf("Reader is nil")
	}

	if reader.Offset() != int64(len(b)) {
		t.Errorf("Expected %d remaining bytes but %d returned instead", len(b), reader.Offset())
	}

	rb := make([]byte, len(b)+1)
	n, err := reader.Read(rb)

	if err != ErrNoContent {
		t.Errorf("Expected ErrNoContent")
	}

	if !bytes.Equal(b, rb[:n]) {
		t.Errorf("Expected read bytes to be equal to source content")
	}

	if reader.Offset() != 0 {
		t.Errorf("Expected %d remaining bytes but %d returned instead", 0, reader.Offset())
	}
}

// TestReaderReverse checks that when reading one byte at a time
// read content should be the reverse of source content
func TestReaderReverse(t *testing.T) {
	b := randomBytes(2048)
	r := &InMemReader{buf: b}

	readSize := 8
	reader, err := NewReader(r, int64(len(b)), readSize)

	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}

	if reader == nil {
		t.Errorf("Reader is nil")
	}

	if reader.Offset() != int64(len(b)) {
		t.Errorf("Expected %d remaining bytes but %d returned instead", len(b), reader.Offset())
	}

	rb := make([]byte, len(b))

	o := len(rb)
	for o > 0 {
		reader.Read(rb[o-1 : o])
		o--
	}

	if !bytes.Equal(b, rb) {
		t.Errorf("Expected read bytes to be equal to source content")
	}

	if reader.Offset() != 0 {
		t.Errorf("Expected %d remaining bytes but %d returned instead", 0, reader.Offset())
	}
}

// TestReaderRecoverEquals checks that reading by chunks of the same size in reverse order
// recover the source content
func TestReaderRecoverEquals(t *testing.T) {
	chunkSpec := randomBytes(100)

	s := 0
	for _, c := range chunkSpec {
		s += int(c)
	}

	b := randomBytes(s)
	r := &InMemReader{buf: b}

	reader, _ := NewReader(r, int64(len(b)), 8)
	rb1 := make([]byte, len(b))

	readUsing(chunkSpec, rb1, reader)

	reader, _ = NewReader(&InMemReader{buf: rb1}, int64(len(rb1)), 4)
	rb2 := make([]byte, len(b))

	readUsing(reverse(chunkSpec), rb2, reader)

	if !bytes.Equal(b, rb2) {
		t.Errorf("Expected read bytes to be equal to orignal content")
	}
}

func reverse(b []byte) []byte {
	r := make([]byte, len(b))

	for i := 0; i < len(r); i++ {
		r[i] = b[len(b)-1-i]
	}

	return r
}

func readUsing(chunkSpec []byte, b []byte, r *Reader) error {
	o := 0
	for _, c := range chunkSpec {
		_, err := r.Read(b[o : o+int(c)])

		if err != nil {
			return err
		}

		o += int(c)
	}
	return nil
}

func TestReadFile(t *testing.T) {
	f, err := ioutil.TempFile(".", "goback_test_file.bin")
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	defer os.Remove(f.Name())

	b := randomBytes(4096)

	n, err := f.Write(b)

	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}

	if n != len(b) {
		t.Errorf("Unexpected error %v", err)
	}

	reader, err := NewFileReader(f)

	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}

	if reader == nil {
		t.Errorf("Reader is nil")
	}

	if reader.Offset() != int64(len(b)) {
		t.Errorf("Expected %d remaining bytes but %d returned instead", len(b), reader.Offset())
	}

	buf := make([]byte, len(b))
	reader.Read(buf)

	if !bytes.Equal(b, buf) {
		t.Errorf("Expected read bytes to be equal to file content")
	}
}

func TestReadAt(t *testing.T) {
	b := randomBytes(1024)
	r := &InMemReader{buf: b}

	readSize := 16
	reader, _ := NewReader(r, int64(len(b)), readSize)

	off := 32
	buf := make([]byte, 1)
	reader.ReadAt(buf, int64(off))

	if !bytes.Equal(b[off-1:off], buf) {
		t.Errorf("Expected read bytes to be equal to source content at offset %d", off)
	}
}
