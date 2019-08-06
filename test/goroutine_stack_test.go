package go_pieces

import (
	"testing"
)

type point struct {
	x int
	y int
}

// String() for implements the fmt.Stringer interface

// if `go test`
// here will got a compile error
// error Sprintf format %s with arg s causes recursive String method call

func (p *point) String() string {
	//return fmt.Sprintf("%s", p) // because Sprintf will call s.String()
	return ""
}

func TestGoroutineStackOverflow(t *testing.T) {
	s := &point{x: 1, y: 2}
	t.Log(s)

	// if run in main package by `go build -v`
	// run will got panic

	// runtime: goroutine stack exceeds 1000000000-byte limit
	//fatal error: stack overflow

}
