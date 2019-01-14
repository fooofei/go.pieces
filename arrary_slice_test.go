package go_pieces

import (
    "fmt"
    "reflect"
)

func ExampleIntPointer() {
    var a [2]int
    var p *int

    p = &a[0]

    fmt.Printf("%v\n", *p)

    for i := 0; i < len(a); i++ {
        fmt.Printf("Element: %d %d\n", i, a[i])
    }
    //output:
    //0
    //Element: 0 0
    //Element: 1 0
}

type IntArray []int

func (a IntArray) Details() {
    for i := 0; i < len(a); i += 1 {
        fmt.Printf("%v of %v=%v\n", i, len(a), a[i])
    }
}
func (a *IntArray) Details2() {
    for i := 0; i < len(*a); i += 1 {
        fmt.Printf("%v of %v=%v\n", i, len(*a), (*a)[i])
    }
}

// cannot use a (type []int) as type []interface {}
func (a IntArray) Breif() string {
    r := fmt.Sprintf("len=%v cap=%v %v", len(a), cap(a), a)
    return r
}

func ExampleBasicSlice() {
    var a []int = []int{1, 2, 3, 4}

    fmt.Printf("reflect.TypeOf(a)= %v\n", reflect.TypeOf(a))
    fmt.Printf("reflect.TypeOf(a).String()==\"[]int\"= %v\n",
        reflect.TypeOf(a).String() == "[]int")
    fmt.Printf("reflect.TypeOf(a).Kind()= %v\n", reflect.TypeOf(a).Kind())
    fmt.Printf("reflect.TypeOf(a).Kind()==reflect.Slice %v\n", reflect.TypeOf(a).Kind() == reflect.Slice)

    fmt.Printf("a= %v\n", a)

    a1 := IntArray(a)
    fmt.Printf("a1.Breif= %v\n", a1.Breif())
    a1.Details()
    IntArray{11, 22}.Details()
    a1.Details2()

    b := IntArray(a[1:])
    fmt.Printf("a[1:].Breif() = %v\n", b.Breif())
    b.Details()
    //output:reflect.TypeOf(a)= []int
    //reflect.TypeOf(a).String()=="[]int"= true
    //reflect.TypeOf(a).Kind()= slice
    //reflect.TypeOf(a).Kind()==reflect.Slice true
    //a= [1 2 3 4]
    //a1.Breif= len=4 cap=4 [1 2 3 4]
    //0 of 4=1
    //1 of 4=2
    //2 of 4=3
    //3 of 4=4
    //0 of 2=11
    //1 of 2=22
    //0 of 4=1
    //1 of 4=2
    //2 of 4=3
    //3 of 4=4
    //a[1:].Breif() = len=3 cap=3 [2 3 4]
    //0 of 3=2
    //1 of 3=3
    //2 of 3=4
}

func ExampleSliceTwice() {
    var a []int = []int{1, 2, 3, 4, 5}
    var b []int = a[0:2]
    var c []int = a[:2]
    var d []int = b[0:2]

    fmt.Printf("a= %v\n", IntArray(a).Breif())
    fmt.Printf("b= %v\n", IntArray(b).Breif())
    fmt.Printf("c= %v\n", IntArray(c).Breif())
    fmt.Printf("d= %v\n", IntArray(d).Breif())

    fmt.Printf("do d[0] = 11\n")
    d[0] = 11

    fmt.Printf("a= %v\n", IntArray(a).Breif())
    fmt.Printf("b= %v\n", IntArray(b).Breif())
    fmt.Printf("c= %v\n", IntArray(c).Breif())
    fmt.Printf("d= %v\n", IntArray(d).Breif())

    //output:a= len=5 cap=5 [1 2 3 4 5]
    //b= len=2 cap=5 [1 2]
    //c= len=2 cap=5 [1 2]
    //d= len=2 cap=5 [1 2]
    //do d[0] = 11
    //a= len=5 cap=5 [11 2 3 4 5]
    //b= len=2 cap=5 [11 2]
    //c= len=2 cap=5 [11 2]
    //d= len=2 cap=5 [11 2]
}

func ExampleFullSliceAppend() {
    var a []int = make([]int, 5, 5)
    fmt.Printf("a= %v\n", IntArray(a).Breif())
    fmt.Printf("append(a)= %v\n", IntArray(append(a, 3)).Breif())
    fmt.Printf("a= %v\n", IntArray(a).Breif())
    fmt.Printf("do a = append(a,4)\n")
    a = append(a, 4)
    fmt.Printf("a= %v\n", IntArray(a).Breif())
    fmt.Printf("do a[0]=11\n")
    a[0] = 11
    fmt.Printf("a= %v\n", IntArray(a).Breif())
    //output:a= len=5 cap=5 [0 0 0 0 0]
    //append(a)= len=6 cap=10 [0 0 0 0 0 3]
    //a= len=5 cap=5 [0 0 0 0 0]
    //do a = append(a,4)
    //a= len=6 cap=10 [0 0 0 0 0 4]
    //do a[0]=11
    //a= len=6 cap=10 [11 0 0 0 0 4]
}

func ExampleEmptySliceAppend() {
    a := make([]int, 0, 5)
    fmt.Printf("a= %v\n", IntArray(a).Breif())
    fmt.Printf("append(a)= %v\n", IntArray(append(a, 3)).Breif())
    fmt.Printf("a= %v\n", IntArray(a).Breif())
    fmt.Printf("do a = append(a,4)\n")
    a = append(a, 4)
    fmt.Printf("a= %v\n", IntArray(a).Breif())
    fmt.Printf("do a[0]=11\n")
    a[0] = 11
    fmt.Printf("a= %v\n", IntArray(a).Breif())
    // output:a= len=0 cap=5 []
    //append(a)= len=1 cap=5 [3]
    //a= len=0 cap=5 []
    //do a = append(a,4)
    //a= len=1 cap=5 [4]
    //do a[0]=11
    //a= len=1 cap=5 [11]
}

func ExampleSliceRange() {
    var a []int
    a = []int{22, 33, 44, 55}

    for i, v := range a {
        fmt.Printf("%v of %v=%v\n", i, len(a), v)
    }
    fmt.Printf("a= %v\n", IntArray(a).Breif())
    //output:0 of 4=22
    //1 of 4=33
    //2 of 4=44
    //3 of 4=55
    //a= len=4 cap=4 [22 33 44 55]
}

func ExampleArray2D() {
    arr := [3][3]int{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}

    // error [...][...]int
    // error [3][...]int
    arr2 := [...][3]int{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}

    fmt.Printf("arr= %v\n", arr)
    fmt.Printf("arr2 = %v\n", arr2)

    for i, v := range arr {
        fmt.Printf("(%v,%v)", i, v)
    }
    fmt.Println("")
    for i, v := range arr {
        for j, w := range v {
            fmt.Printf("(%v,%v,%v)", i, j, w)
        }
    }
    fmt.Println("")
    //output:arr= [[1 2 3] [4 5 6] [7 8 9]]
    //arr2 = [[1 2 3] [4 5 6] [7 8 9]]
    //(0,[1 2 3])(1,[4 5 6])(2,[7 8 9])
    //(0,0,1)(0,1,2)(0,2,3)(1,0,4)(1,1,5)(1,2,6)(2,0,7)(2,1,8)(2,2,9)
}

func ExampleOutOfRangeSlice() {
    arr := [...]int{1, 2, 3}
    fmt.Printf("arr= %v\n", arr)
    fmt.Printf("arr[2:]=%v\n", arr[2:])
    // panic: runtime error: slice bounds out of range
    //fmt.Printf("arr[5:]=%v\n", arr[2:][5:])
    _ = arr
    //output:arr= [1 2 3]
    //arr[2:]=[3]
}

func ExampleDeleteFromSlice() {
    var arr [3]int = [...]int{1, 2, 3}

    var s = arr[:]

    fmt.Printf("s= %v\n", IntArray(s).Breif())

    popFront := append(s[1:])
    fmt.Printf("popFront= %v\n", IntArray(popFront).Breif())
    popBack := append([]int{}, s[:len(s)-1]...)
    fmt.Printf("popBack= %v\n", IntArray(popBack).Breif())
    popBack2 := append(make([]int, 0), s[:len(s)-1]...)
    fmt.Printf("popBack2= %v\n", IntArray(popBack2).Breif())
    //output:s= len=3 cap=3 [1 2 3]
    //popFront= len=2 cap=2 [2 3]
    //popBack= len=2 cap=2 [1 2]
    //popBack2= len=2 cap=2 [1 2]
}
