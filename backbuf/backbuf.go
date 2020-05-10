package backbuf

import (
	"fmt"
)

const minSize = 1

// Buffer implements a circular byte buffer, data is read from left to right
// but chucks are read from right to left. Unread data is never overwriten
type Buffer struct {
	buffer []byte

	wpos int
	rpos int
	full bool
}

// New creates a Buffer with a capacity of `size´ bytes
func New(size int) (*Buffer, error) {
	if size < minSize {
		return nil, fmt.Errorf("Size must be greater than %d", minSize)
	}

	b := &Buffer{buffer: make([]byte, size)}
	b.Reset()
	return b, nil
}

// Size returns the capacity of the buffer
func (cb *Buffer) Size() int {
	return len(cb.buffer)
}

// Reset makes the buffer fully available
func (cb *Buffer) Reset() {
	cb.wpos = len(cb.buffer)
	cb.rpos = cb.wpos
	cb.full = false
}

// WriteAvailability returns the number of bytes that can be writen
// without overwriting unread data
func (cb *Buffer) WriteAvailability() int {
	if cb.full {
		return 0
	}

	if cb.wpos <= cb.rpos {
		return len(cb.buffer) - cb.rpos + cb.wpos
	}

	return cb.wpos - cb.rpos
}

// ReadAvailability returns the number of bytes that can be read
func (cb *Buffer) ReadAvailability() int {
	return len(cb.buffer) - cb.WriteAvailability()
}

// Write stores bytes into the buffer without overriding unread data
func (cb *Buffer) Write(buf []byte) (n int, err error) {
	if buf == nil || len(buf) == 0 {
		return 0, fmt.Errorf("Input buffer is null or empty")
	}

	toWrite := min(cb.WriteAvailability(), len(buf))

	if toWrite == 0 {
		return 0, fmt.Errorf("Buffer is full")
	}

	if cb.wpos <= cb.rpos {
		if toWrite < cb.wpos {
			copy(cb.buffer[cb.wpos-toWrite:cb.wpos], buf[:toWrite])
			cb.wpos -= toWrite
		} else {
			copy(cb.buffer[:cb.wpos], buf[:cb.wpos])
			copy(cb.buffer[len(cb.buffer)-toWrite+cb.wpos:len(cb.buffer)], buf[cb.wpos:toWrite])
			cb.wpos = len(cb.buffer) - toWrite + cb.wpos
		}
	} else {
		copy(cb.buffer[cb.wpos-toWrite:cb.wpos], buf[:toWrite])
		cb.wpos -= toWrite
	}

	cb.full = cb.wpos == cb.rpos
	return toWrite, nil
}

// Read reads up to len(buf) bytes
func (cb *Buffer) Read(buf []byte) (int, error) {
	if buf == nil || len(buf) == 0 {
		return 0, fmt.Errorf("Target buffer is null or empty")
	}

	toRead := min(len(buf), cb.ReadAvailability())

	lBytes := min(cb.rpos, toRead)
	copy(buf[toRead-lBytes:toRead], cb.buffer[cb.rpos-lBytes:cb.rpos])

	if toRead > lBytes {
		rBytes := toRead - lBytes
		copy(buf[:rBytes], cb.buffer[len(cb.buffer)-rBytes:len(cb.buffer)])
		cb.rpos = len(cb.buffer) - rBytes
	} else {
		cb.rpos -= lBytes
	}

	cb.full = false
	return toRead, nil
}

// ReadByte returns the first unread byte
func (cb *Buffer) ReadByte() (byte, error) {
	buf := make([]byte, 1)
	n, err := cb.Read(buf)

	if err != nil {
		return 0, err
	}

	if n == 0 {
		return 0, fmt.Errorf("Buffer is empty")
	}

	return buf[0], nil
}

// Omit updates read position as if n bytes were read
func (cb *Buffer) Omit(n int) error {
	if n < 1 {
		return fmt.Errorf("Positive number required")
	}

	if cb.ReadAvailability() < n {
		return fmt.Errorf("Not enough unread data")
	}

	if cb.rpos <= n {
		cb.rpos = len(cb.buffer) - n + cb.rpos
	} else {
		cb.rpos -= n
	}

	cb.full = false
	return nil
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
