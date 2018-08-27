package main

import "fmt"

func addr() func(int) int {
	var sum int =0
	return func(x int) int {
		sum += x
		return sum
	}
}

/*
pos(0)=0
neg(-2*0)=0
pos(1)=1
neg(-2*1)=-2
pos(2)=3
neg(-2*2)=-6
pos(3)=6
neg(-2*3)=-12
pos(4)=10
neg(-2*4)=-20
pos(5)=15
neg(-2*5)=-30
pos(6)=21
neg(-2*6)=-42
pos(7)=28
neg(-2*7)=-56
pos(8)=36
neg(-2*8)=-72
pos(9)=45
neg(-2*9)=-90

 */
func testAddr(){
	pos,neg := addr(), addr()
	for i:=0;i<10;i+=1{
		fmt.Printf("pos(%v)=%v\n",i, pos(i))
		fmt.Printf("neg(-2*%v)=%v\n",i, neg(-2*i))
	}
}

func main6(){
	testAddr()
}