package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"
	"math"
	"runtime"
)

func Afunc(x int, y int) int  {
	return x+y
}

// 使用了 runtime 中的 const
func switchFunc(){
	fmt.Printf("ARCH=%s\n", runtime.GOARCH);
	fmt.Printf("OS=%v\n", runtime.GOOS)

	switch runtime.GOOS {
	case "darwin":
		fmt.Printf("Platform macOS\n")
	default:
		fmt.Printf("Platform unknown\n")
	}
}

// 定义一个 struct
type myStruct struct {
	x int
	y int
}

// 给这个struct 增加一个成员函数
func (self *myStruct)String() (r string) {
	r = fmt.Sprintf("<myStruct>(%v,%v)", self.x, self.y)
	return r
}

// 使用这个 struct
func structFunc(){
	var a myStruct=myStruct{1,1}

	fmt.Printf("value=%v\n", a.String())
}

func main(){
	var arg string
	var i int
	fmt.Fprintf(os.Stdout, "Arg count=%d\n", len(os.Args))
	for i=0;i<len(os.Args);i++{
		arg = os.Args[i]
		fmt.Fprintf(os.Stdout, "%s\n" , arg)
	}

	rand.Seed(time.Now().Unix())
	fmt.Fprintf(os.Stdout,"number= %d\n", rand.Intn(10))
	fmt.Fprintf(os.Stdout, "Phi= %f\n", math.Phi)
	fmt.Fprintf(os.Stdout, "Pi= %f\n", math.Pi)
	fmt.Fprintf(os.Stdout, "Afunc(1,11)=%d\n", Afunc(1,11))
	switchFunc()
	structFunc()
}

