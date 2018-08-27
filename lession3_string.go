package main

import (
	"fmt"
	"strings"
)

func main(){
	var a string = "  foo bar  baz   "
	fmt.Printf("Fields are: %q\n", strings.Fields(a))
}
