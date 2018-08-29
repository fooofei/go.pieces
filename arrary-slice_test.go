package main

import (
    "fmt"
    "reflect"
)

func sliceBreif(a []int) string{
    r:= fmt.Sprintf("len=%v,cap=%v,%v", len(a), cap(a), a)
    return r
}

func ExamplePointerAndSlice()  {
    var a [2]int
    var p *int

    p = &a[0]

    fmt.Printf("%v\n", *p)

    for i := 0 ; i<len(a) ; i++ {
        fmt.Printf("Element: %d %d\n", i, a[i])
    }
    //output:
    //0
    //Element: 0 0
    //Element: 1 0
}

type IntArray []int

func (a IntArray)Details(){
    for i:=0;i<len(a);i+=1{
        fmt.Printf("%v of %v=%v\n",i,len(a),a[i])
    }
}
func (a * IntArray)Details2(){
    for i:=0;i<len(*a);i+=1{
        fmt.Printf("%v of %v=%v\n",i,len(*a),(*a)[i])
    }
}

func ExampleBasicSlice(){
    var a []int = []int{1,2,3,4}

    fmt.Printf("%v %v\n",reflect.TypeOf(a),a)

    var a1 IntArray = IntArray(a)
    a1.Details()
    IntArray{11,22}.Details()

    a1.Details2()

    var b []int = a[1:]
    for i:=0;i<len(b);i+=1{
        fmt.Printf("element %v\n", b[i])
    }
    //output:
    //[]int [1 2 3 4]
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
    //element 2
    //element 3
    //element 4
}


func ExampleSliceTwice(){
    var a []int = []int{1,2,3,4,5}
    var b []int = a[0:2]
    var c []int = a[:2]
    var d []int = b[0:2]

    fmt.Printf("a=%v b=%v c=%v d=%v\n", a,b,c,d)
    fmt.Printf("a=%v,b=%v,c=%v,d=%v\n",sliceBreif(a),sliceBreif(b),sliceBreif(c),sliceBreif(d))
    d[0] = 11
    fmt.Printf("a=%v b=%v c=%v d=%v\n", a,b,c,d)
    fmt.Printf("len(a)=%v cap(a)=%v len(d)=%v cap(d)=%v \n", len(a), cap(a),
        len(d), cap(d))
    //output:
    //a=[1 2 3 4 5] b=[1 2] c=[1 2] d=[1 2]
    //a=len=5,cap=5,[1 2 3 4 5],b=len=2,cap=5,[1 2],c=len=2,cap=5,[1 2],d=len=2,cap=5,[1 2]
    //a=[11 2 3 4 5] b=[11 2] c=[11 2] d=[11 2]
    //len(a)=5 cap(a)=5 len(d)=2 cap(d)=5
}

func ExampleSliceBreif(){
    var a []int = make([]int, 5,5)
    fmt.Printf("a=%v\n",sliceBreif(a))
    fmt.Printf("append(a)=%v\n", sliceBreif(append(a,3)))
    fmt.Printf("a=%v\n",sliceBreif(a))
    a = append(a,4)
    fmt.Printf("after a=append(a) a=%v\n",sliceBreif(a))
    a[0]=11
    fmt.Printf("after a[0]=11,a=%v\n",sliceBreif(a))
    fmt.Println("")

    var b []int = make([]int, 0, 5)
    fmt.Printf("b=%v\n", sliceBreif(b))
    fmt.Printf("append(b)=%v\n", sliceBreif(append(b, 3)))
    fmt.Printf("b=%v\n",sliceBreif(b))
    b = append(b,4)
    fmt.Printf("after b=append(b) b=%v\n", sliceBreif(b))
    // output:
    //a=len=5,cap=5,[0 0 0 0 0]
    //append(a)=len=6,cap=10,[0 0 0 0 0 3]
    //a=len=5,cap=5,[0 0 0 0 0]
    //after a=append(a) a=len=6,cap=10,[0 0 0 0 0 4]
    //after a[0]=11,a=len=6,cap=10,[11 0 0 0 0 4]
    //
    //b=len=0,cap=5,[]
    //append(b)=len=1,cap=5,[3]
    //b=len=0,cap=5,[]
    //after b=append(b) b=len=1,cap=5,[4]
}


func ExampleSliceRange(){
    var a []int
    a=[]int{22,33,44,55}

    for i,v := range a{
        fmt.Printf("%v of %v=%v\n",i,len(a),v)
    }
    sliceBreif(a)
    //output:
    //0 of 4=22
    //1 of 4=33
    //2 of 4=44
    //3 of 4=55
}


func ExampleArray2D(){
    var arr [3][3]int
    arr = [3][3]int{{1,2,3},{4,5,6},{7,8,9}}

    // error [...][...]int
    // error [3][...]int
    var arr2 [3][3]int =[...][3]int{{1,2,3},{4,5,6},{7,8,9}}

    fmt.Printf("%v\n",arr)
    fmt.Printf("%v\n", arr2)

    for i,v := range arr{
        fmt.Printf("(%v,%v)",i,v)
    }
    fmt.Println("")
    for i,v := range arr{
        for j,w := range v{
            fmt.Printf("(%v,%v,%v)",i,j,w)
        }
    }
    fmt.Println("")
    //output:
    //[[1 2 3] [4 5 6] [7 8 9]]
    //[[1 2 3] [4 5 6] [7 8 9]]
    //(0,[1 2 3])(1,[4 5 6])(2,[7 8 9])
    //(0,0,1)(0,1,2)(0,2,3)(1,0,4)(1,1,5)(1,2,6)(2,0,7)(2,1,8)(2,2,9)
}


func ExampleOutOfRangeSlice(){
    var arr [3]int = [...]int{1,2,3}
    fmt.Printf("%v\n",arr)
    fmt.Printf("arr[2:]=%v\n", arr[2:])
    // out range
    //fmt.Printf("arr[5:]=%v\n", arr[2:][5:])
    //output:
    //[1 2 3]
    //arr[2:]=[3]
}


func ExampleDeleteFromSlice(){
    var arr [3]int = [...]int {1,2,3}

    var s = arr[:]

    fmt.Printf("s=%v\n",s)

    fmt.Printf("pop front=%v\n", append(s[1:]))
    fmt.Printf("pop back=%v\n", append([]int{},s[:len(s)-1]...))
    fmt.Printf("pop back=%v\n", append(make([]int,0),s[:len(s)-1]...))
    //output:
    //s=[1 2 3]
    //pop front=[2 3]
    //pop back=[1 2]
    //pop back=[1 2]
}