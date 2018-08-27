package main

import (
  "fmt"
  "reflect"
)

func sliceView(a []int) (r string){
  r = fmt.Sprintf("len=%v cap=%v %v", len(a), cap(a),a)
  return r
}

func normalFunc(){
  var a [2]int
  var p *int

  p = &a[0]

  fmt.Printf("%v\n", *p)

  for i := 0 ; i<len(a) ; i++ {
    fmt.Printf("Element: %d %d\n", i, a[i])
  }
}

type int_array_type []int
func (a  int_array_type) view(){
  for i:=0;i<len(a);i+=1{
    fmt.Printf("[%v/%v] %v\n", i, len(a), a[i])
  }

}

// 这两个方法的区别是看你要不要在方法里修改这个类实例的值
/*
func (a * int_array_type)view1(){
  for i:=0;i<len(*a);i+=1{
    fmt.Printf("[%v/%v] %v\n", i, len(*a), (*a)[i])
  }
}
*/

func sliceFunc(){
  var a []int = []int{1,2,3,4}

  fmt.Printf("%v %v\n",reflect.TypeOf(a),a)

  var a1 int_array_type = int_array_type(a)
  a1.view()
  int_array_type{11,22}.view()

  var b []int = a[1:]

  for i:=0;i<len(b);i+=1{
    fmt.Printf("element %v\n", b[i])
  }
}

/*
a=[1 2 3 4 5] b=[1 2] c=[1 2] d=[1 2]
a=[11 2 3 4 5] b=[11 2] c=[11 2] d=[11 2]
 */
func sliceTwice(){
  var a []int = []int{1,2,3,4,5}
  var b []int = a[0:2]
  var c []int = a[:2]
  var d []int = b[0:2]

  fmt.Printf("a=%v b=%v c=%v d=%v\n", a,b,c,d)
  d[0] = 11
  fmt.Printf("a=%v b=%v c=%v d=%v\n", a,b,c,d)

  fmt.Printf("len(a)=%v cap(a)=%v len(d)=%v cap(d)=%v \n", len(a), cap(a),
    len(d), cap(d))
}

/*
a=len=5 cap=5 [0 0 0 0 0]
append(a)=len=6 cap=10 [0 0 0 0 0 3]
a=len=5 cap=5 [0 0 0 0 0]
b=len=0 cap=5 []
append(b)=len=1 cap=5 [3]
b=len=0 cap=5 []
 */
 // 说明如果不把 append 返回值保存下来就白 append 了， append 不是在原 array 上操作
func sliceAppend(){
  var a []int = make([]int,5,5)
  fmt.Printf("a=%v\n", sliceView(a))

  fmt.Printf("append(a)=%v\n", sliceView(append(a,3)))
  fmt.Printf("a=%v\n",sliceView(a))

  var b []int = make([]int, 0,5)
  fmt.Printf("b=%v\n", sliceView(b))
  fmt.Printf("append(b)=%v\n", sliceView(append(b,3)))
  fmt.Printf("b=%v\n", sliceView(b))

}

func sliceRange(){
  var a []int = []int{22,33,44,55}

  for i,v := range a{
    fmt.Printf("[%v/%v]%v\n", i,len(a),v)
  }
}
func main() {
  sliceRange()
}
