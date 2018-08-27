package main

import "fmt"

func do(i interface{}) {
	switch v := i.(type) {
	case int:
		fmt.Printf("Twice %v is %v\n", v, v*2)
	case string:
		fmt.Printf("%q is %v bytes long\n", v, len(v))
	default:
		fmt.Printf("I don't know about type %T!\n", v)
	}
}

func testTypeSwitch(){
	do(21)
	do("hello")
	do(true)
}

func testType1(){
	var a int = 3
	var c interface{} = a

	b,ok := c.(float64)
	fmt.Printf("type(a)=%T b=%v ok=%v\n",a,b,ok)
}


func main() {
	testType1()
}
