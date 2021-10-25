package buffer

import (
	"sync"
)

var pool = &sync.Pool{
	New: func() interface{} {
		return &Buffer{
			buf: make([]byte, 0, 256),
		}
	},
}

// Buffer is similar to a strings.Builder, but is pooled and reusable.
// To create a Buffer use `New()` and when work its done call Buffer.Release()
type Buffer struct {
	buf []byte
}

func New() (b *Buffer) {
	return pool.Get().(*Buffer)
}

// WriteString appends the contents of s to b's buffer.
// It returns the length of s and a nil error.
func (b *Buffer) WriteString(s string) (int, error) {
	b.buf = append(b.buf, s...)
	return len(b.buf), nil
}

// Write appends the contents of p to b's buffer.
// Write always returns len(p), nil.
func (b *Buffer) Write(p []byte) (int, error) {
	b.buf = append(b.buf, p...)
	return len(p), nil
}

// WriteByte appends the byte c to b's buffer.
// The returned error is always nil.
func (b *Buffer) WriteByte(c byte) error {
	b.buf = append(b.buf, c)
	return nil
}

// String returns the accumulated string.
func (b *Buffer) String() string {
	return string(b.buf)
}

// Len returns the number of accumulated bytes; b.Len() == len(b.String()).
func (b *Buffer) Len() int { return len(b.buf) }

// Cap returns the capacity of the builder's underlying byte slice. It is the
// total space allocated for the string being built and includes any bytes
// already written.
func (b *Buffer) Cap() int { return cap(b.buf) }

// Release releases the buffer making it available for reutilization
func (b *Buffer) Release() {
	b.buf = b.buf[:0]
	pool.Put(b)
}
