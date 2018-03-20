package main

import (
	"fmt"
)

func testInterface(a interface{}) (ret interface{}, er error) {

	if s,ok := a.(string);ok{
		fmt.Printf("We got %T:%v in testInterface\n",s, s)
		return "We return a string",nil
	} else if i,ok := a.(int);ok{
		fmt.Printf("We got %T:%v in testInterface\n",i,i)
		return -1,nil
	}
	return nil,fmt.Errorf("not found the right type")

}

func main() {

	v1,_ := testInterface("nnn")
	fmt.Printf("v1=%v\n",v1)

	v2,_:= testInterface(1)
	fmt.Printf("v2=%v\n",v2)

	v3,_:=testInterface(nil)
	fmt.Printf("v3=%v\n",v3)

	// 测试返回值个数
	//  在 testInterface 中，返回值有 2 个，
	//    我们调用这个函数也必须接受 2 个返回值
	//  发现类型转换 有时候是 2 个 有时候 1 个也可以
	//  咨询他人说这是语法问题，可以这么做的有两个 1 类型转换可以 2 map的查找也可以
	v4 := v1.(string)
	fmt.Printf("%v\n",v4)

}