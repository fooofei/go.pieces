package main


import "fmt"

type S struct {
	a, b int
}

// String implements the fmt.Stringer interface
// compile error
// # github.com/fooofei/go_pieces
//./goroutinestack_test.go:12: Sprintf format %s with arg s causes recursive String method call
func (s *S) String() string {
	//return fmt.Sprintf("%s", s) // Sprintf will call s.String()
	return ""
}


func ExampleGoroutineStackOverflow() {
	s := &S{a: 1, b: 2}
	fmt.Println(s)
	// runtime: goroutine stack exceeds 1000000000-byte limit
	//fatal error: stack overflow
}