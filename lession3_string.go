package main

import (
	"fmt"
	"strings"
)

func main7(){
	var a string = "  foo bar  baz   "
	fmt.Printf("Fields are: %q\n", strings.Fields(a))
}
