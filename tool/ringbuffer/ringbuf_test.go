package main

import (
	"testing"

	"gotest.tools/assert"
)

func TestRingbuf1(t *testing.T) {
	rb := New(3)
	s := "hellow"
	n, _ := rb.Write([]byte(s))
	assert.Equal(t, n, len(s))

	b := make([]byte, 100)
	n, _ = rb.Read(b)
	assert.Equal(t, n, 3)
	assert.Equal(t, string(b[:n]), s[len(s)-rb.Cap():])
}
