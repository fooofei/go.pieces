package main

import (
    "fmt"
    "path/filepath"
    "runtime"
)

func GetCurrentDir() (executablePath string) {
    _, callerFile, _, _ := runtime.Caller(0)
    executablePath = filepath.Dir(callerFile)
    return executablePath
}

func ExampleGetCurDir(){
    var cur = GetCurrentDir()
    var name = filepath.Base(cur)
    fmt.Printf("getcurdir=%v\n",name)
    //output:
    // getcurdir=go_pieces
}



