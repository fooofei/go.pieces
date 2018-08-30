package main

import "fmt"



func addr() func(int) int {
	var sum int =0
	return func(x int) int {
		sum += x
		return sum
	}
}


func ExampleClosureFunc(){
    pos,neg := addr(),addr()
    fmt.Printf("idx,pos(idx),neg(-2*idx)\n")
    for i:=0;i<10;i+=1{
        fmt.Printf("%v %v  %v\n", i,pos(i),neg(-2*i))
    }
    //output:
    //idx,pos(idx),neg(-2*idx)
    //0 0  0
    //1 1  -2
    //2 3  -6
    //3 6  -12
    //4 10  -20
    //5 15  -30
    //6 21  -42
    //7 28  -56
    //8 36  -72
    //9 45  -90
}

func change(v func(int)int) int{
    return v(11)
}

// closure func as param will be reference
func ExampleClosureFuncParam(){
    var a = addr()
    fmt.Printf("a(1)=%v\n",a(1))
    fmt.Printf("a(2)=%v\n",a(2))
    fmt.Printf("a(0)=%v\n",a(0))
    fmt.Printf("change(a)=%v\n", change(a))
    fmt.Printf("a(0)=%v\n", a(0))
    //output:
    //a(1)=1
    //a(2)=3
    //a(0)=3
    //change(a)=14
    //a(0)=14
}

func ExampleClosureFuncCopy(){
    var a = addr()
    fmt.Printf("a(1)=%v\n",a(1))
    fmt.Printf("a(2)=%v\n",a(2))
    fmt.Printf("a(0)=%v\n",a(0))

    var b = a
    fmt.Printf("b(0)=%v\n",b(0))
    fmt.Printf("b(1)=%v\n",b(1))
    fmt.Printf("a(0)=%v\n",a(0))
    fmt.Printf("b(0)=%v\n",b(0))
    //output:
    //a(1)=1
    //a(2)=3
    //a(0)=3
    //b(0)=3
    //b(1)=4
    //a(0)=4
    //b(0)=4
}