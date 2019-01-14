package go_pieces

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


func aInterface(a interface{}) (interface{}, error){
	if s,ok :=a.(string); ok{
		return fmt.Sprintf("we got a string %v\n", s), nil
	} else if i,ok := a.(int);ok {
		return fmt.Sprintf("we got an int %v\n", i), nil
	}
	return nil, fmt.Errorf("unknown type %v\n", a)

}

func ExampleInterface(){

	a,err := aInterface("yes")
	fmt.Printf("%v,%v\n", a, err)

	b,err := aInterface(10)
	fmt.Printf("%v,%v\n", b,err)
	fmt.Println(aInterface(true))
	// output:
	//we got a string yes
	//,<nil>
	//we got an int 10
	//,<nil>
	//<nil> unknown type true
}

func ExampleTypeCastCheck(){

	a:= "a string"

	// a.(string) error  (non-interface type string on left)
	s1 := interface{}(a).(string)
	fmt.Println(s1)

	s2,ok :=interface{}(a).(string)
	fmt.Println(s2,ok)
	// output:
	//a string
	//a string true
}

