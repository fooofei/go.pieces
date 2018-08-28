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

