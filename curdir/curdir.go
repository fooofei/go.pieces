package main

import (
    "fmt"
    "os"
    "path/filepath"
    "runtime"
)

func CurrentDirOfRuntime0() () {
    // only compile the .go source file directory
    // not the executable file
    // when go build xx.go
    // ./xx
    // the output dir is wrong
    _, callerFile, _, _ := runtime.Caller(0)
    fmt.Printf("runtime.Caller(0)= \n")
    fmt.Printf("  file= %v\n", callerFile)
    fmt.Printf("  dir= %v\n", filepath.Dir(callerFile))
}

func OsExecutable() {
    // the executable file path
    // when go run xx.go
    // the output path is a temp path
    cFile,_ := os.Executable()
    fmt.Printf("os.Executable=\n")
    fmt.Printf("  file= %v\n", cFile)
    fmt.Printf("  dir= %v\n", filepath.Dir(cFile))
}


func main(){
    CurrentDirOfRuntime0()
    OsExecutable()
}
