package backbuf

import (
	"bytes"
	"testing"
)

func TestBufferCreation(t *testing.T) {
	const size = minSize + 1

	cb, err := New(size)

	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}

	if cb == nil {
		t.Errorf("No buffer returned")
	}

	if cb.Size() != size {
		t.Errorf("Expected size %d but returned %d", size, cb.Size())
	}

	if cb.WriteAvailability() != size {
		t.Errorf("Expected full availability but returned %d", cb.WriteAvailability())
	}

	if cb.ReadAvailability() != 0 {
		t.Errorf("Expected none availability but returned %d", cb.ReadAvailability())
	}
}

func TestBufferInvalidSize(t *testing.T) {
	_, err := New(minSize - 1)

	if err == nil {
		t.Errorf("Expected error to be returned when size is lower than %d", minSize)
	}
}

func TestBufferWriteReadFlow(t *testing.T) {
	const size = 5

	cb, err := New(size)

	wa := cb.WriteAvailability()

	if wa != size {
		t.Errorf("Expected full availability but returned %d", wa)
	}

	d := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	n, err := cb.Write(d)

	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}

	if n != min(size, len(d)) {
		t.Errorf("Less than %d were written", min(size, len(d)))
	}

	if cb.WriteAvailability() != wa-n {
		t.Errorf("Expected write availability %d but returned %d instead", wa-n, cb.WriteAvailability())
	}

	if cb.ReadAvailability() != n {
		t.Errorf("Expected read availability %d but returned %d instead", n, cb.ReadAvailability())
	}

	r := make([]byte, cb.ReadAvailability())

	_, err = cb.Read(r)

	if cb.ReadAvailability() != 0 {
		t.Errorf("Expected read availability %d but returned %d instead", 0, cb.ReadAvailability())
	}

	if !bytes.Equal(r, d[:len(r)]) {
		t.Errorf("Expected read bytes to be equal to prefix of input bytes")
	}
}

func TestBufferFlow(t *testing.T) {
	const size = 5

	cb, err := New(size)

	wa := cb.WriteAvailability()

	if wa != size {
		t.Errorf("Expected full availability but returned %d", wa)
	}

	d := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	n, err := cb.Write(d)

	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}

	if n != min(size, len(d)) {
		t.Errorf("Less than %d were written", min(size, len(d)))
	}

	if cb.WriteAvailability() != wa-n {
		t.Errorf("Expected write availability %d but returned %d instead", wa-n, cb.WriteAvailability())
	}

	if cb.ReadAvailability() != n {
		t.Errorf("Expected read availability %d but returned %d instead", n, cb.ReadAvailability())
	}

	r := make([]byte, 2)

	_, err = cb.Read(r)

	n, err = cb.Write(d[size:])

	r = make([]byte, cb.ReadAvailability()-1)

	_, err = cb.Read(r)

	r = make([]byte, 0)
}
