package main

import (
	"fmt"
)


type SimpleStruct struct {
	x int
	y int
}

func (self * SimpleStruct) String() string {
	return fmt.Sprintf("<SimpleStruct> (x=%v,y=%v)", self.x, self.y)
}

func (self * SimpleStruct) method1 (){
	self.x *= 2
	self.y *=2
}

func (self  SimpleStruct) method2(){
	self.x *= 2
	self.y *= 2
}


func ExampleSimpleStruct1(){
	var a SimpleStruct = SimpleStruct{11,22}
	fmt.Printf("a=%v\n",a.String())
	a.method1()
	fmt.Printf("after method1 a=%v\n",a.String())
	//output:
	// a=<SimpleStruct> (x=11,y=22)
	//after method1 a=<SimpleStruct> (x=22,y=44)
}

func ExampleSimpleStruct2(){
	a := SimpleStruct{11,22}
	fmt.Printf("a=%v\n",a.String())
	a.method2()
	fmt.Printf("after method2 a=%v\n",a.String())
	//output:
	//a=<SimpleStruct> (x=11,y=22)
	//after method2 a=<SimpleStruct> (x=11,y=22)
}


func ExampleNilStruct(){
	var a *SimpleStruct

	a.method1()

	// SIGSEGV error
	//fmt.Printf("a=%v\n",a)
	//output:
}