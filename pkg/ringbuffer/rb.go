package ringbuffer

// ringbuffer from https://github.com/mozilla-services/heka/blob/dev/ringbuf/ringbuf.go

type Buffer struct {
	buf   []byte
	start int
	size  int
}

// New creates a new ring buffer of the given size.
func New(size int) *Buffer {
	return &Buffer{make([]byte, size), 0, 0}
}

// Write appends all the bytes in b to the buffer, looping and overwriting
// as needed, while incrementing the start to point to the start of the
// buffer.
func (r *Buffer) Write(b []byte) (int, error) {
	written := len(b)
	for len(b) > 0 {
		start := (r.start + r.size) % len(r.buf)

		// Copy as much as we can from where we are
		n := copy(r.buf[start:], b)
		b = b[n:]

		// Are we already at capacity? Move the start an appropriate
		// distance forward depending on how much we copied.
		if r.size >= len(r.buf) {
			if n <= len(r.buf) {
				r.start += n
				if r.start >= len(r.buf) {
					r.start = 0
				}
			} else {
				r.start = 0
			}
		}
		r.size += n
		// Size can't exceed the capacity
		if r.size > cap(r.buf) {
			r.size = cap(r.buf)
		}
	}
	return written, nil
}

// Read reads as many bytes as possible from the ring buffer into
// b. Returns the number of bytes read.
func (r *Buffer) Read(b []byte) (int, error) {
	read := 0
	size := r.size
	start := r.start
	for len(b) > 0 && size > 0 {
		end := start + size
		if end > len(r.buf) {
			end = len(r.buf)
		}
		n := copy(b, r.buf[start:end])
		size -= n
		read += n
		b = b[n:]
		start = (start + n) % len(r.buf)
	}
	return read, nil
}

func (r *Buffer) Size() int {
	return r.size
}

func (r *Buffer) Cap() int {
	return len(r.buf)
}
