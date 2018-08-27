package main

import "fmt"

type  myStruct struct {
	x int
	y int
}

func(self * myStruct)string() string{
	return fmt.Sprintf("<myStruct> (x=%v,y=%v)", self.x, self.y)
}

func (self *myStruct)method1(){
	self.x *=2
	self.y *=2
}

func (self myStruct)method2(){
	self.x *=2
	self.y *=2
}
/*
a=<myStruct> (x=11,y=22)
after method1 a=<myStruct> (x=22,y=44)
a=<myStruct> (x=11,y=22)
after method2 a=<myStruct> (x=11,y=22)

 */
func test1(){
	var a myStruct=myStruct{11,22}
	fmt.Printf("a=%v\n",a.string())
	a.method1()
	fmt.Printf("after method1 a=%v\n",a.string())
}
func test2(){
	var a myStruct=myStruct{11,22}
	fmt.Printf("a=%v\n",a.string())
	a.method2()
	fmt.Printf("after method2 a=%v\n",a.string())
}

func main(){
	test1()
	test2()
}
