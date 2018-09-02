package main

import (
    "fmt"
    "path"
    "reflect"
    "runtime"
)

func GetFuncName(f interface{}) string{
    var a = reflect.ValueOf(f)
    var b = a.Pointer()
    var c = runtime.FuncForPC(b)
    var fullpath = c.Name()
    var moduleName = path.Base(fullpath)
    var ext = path.Ext(moduleName)
    return ext[1:]
    }


func ExampleGetFuncName(){
    var name = GetFuncName(ExampleGetFuncName)
    fmt.Printf("%v name=%v\n",name,name)
    //output:ExampleGetFuncName name=ExampleGetFuncName
}
