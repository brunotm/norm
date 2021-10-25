package buffer

import (
	"testing"
)

func TestNew(t *testing.T) {
	buf := New()
	defer buf.Release()

	if buf.Cap() == 0 {
		t.Errorf("buffer capacity expected to be greater than 0")
	}

	if buf.Len() > 0 {
		t.Errorf("buffer length expected to be 0")
	}
}

func TestBuffer_WriteString(t *testing.T) {
	buf := New()
	defer buf.Release()

	s := "this is a string"
	_, _ = buf.WriteString(s)

	if buf.String() != s {
		t.Errorf("expected: %s, got: %s", s, buf.String())
	}
}

func TestBuffer_Write(t *testing.T) {
	buf := New()
	defer buf.Release()

	s := []byte("this is a string")
	_, _ = buf.Write(s)

	if buf.String() != string(s) {
		t.Errorf("expected: %s, got: %s", string(s), buf.String())
	}
}

func TestBuffer_WriteByte(t *testing.T) {
	buf := New()
	defer buf.Release()

	s := byte(1)
	_ = buf.WriteByte(s)

	if buf.String() != string(s) {
		t.Errorf("expected: %s, got: %s", string(s), buf.String())
	}
}
