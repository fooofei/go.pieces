package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"path"
	"runtime"
	"time"
)


// 定义一个 struct
type myStruct struct {
	x int
	y int
}

// 给这个struct 增加一个成员函数
func (self *myStruct)String() (string) {
	r := fmt.Sprintf("<myStruct>(%v,%v)", self.x, self.y)
	return r
}

func ExampleMyStruct(){
	var a myStruct=myStruct{1,1}

	fmt.Printf("mystring=%v\n", a.String())
	//output:mystring=<myStruct>(1,1)
}

func ExampleCmdArgs(){
	argsLen := len(os.Args)

	fmt.Printf("argsLen:%v\n", argsLen)

	if 0< argsLen{
		exeName := path.Base(os.Args[0])
		fmt.Printf("exeName=%v\n", exeName)
	}

	//output:argsLen:2
	//exeName=sub.test
}

func ExampleSomeConstants(){
	rand.Seed(time.Now().Unix())
	n := rand.Intn(10)
	b := (n>=0 && n <10)
	fmt.Printf("randNum=[0,10) %v\n", b)
	fmt.Printf("Phi=%.3f\n", math.Phi)
	fmt.Printf("Pi=%.3f\n", math.Pi)
	fmt.Printf("GOOS=%v\n",runtime.GOOS)
	fmt.Printf("GOARCH=%v\n",runtime.GOARCH)
	// output:
	//randNum=[0,10) true
	//Phi=1.618
	//Pi=3.142
	//GOOS=darwin
	//GOARCH=amd64
}


