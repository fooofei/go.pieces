package main

import (
    "fmt"
    "net"
)

func main(){

    names, err := net.LookupAddr("")
    if err != nil {
        panic(err)
    }
    if len(names) == 0 {
        fmt.Printf("no record")
    }
    for _, name := range names {
        fmt.Printf("%s\n", name)
    }

    r,err := net.LookupIP("")
    if err != nil {
        panic(err)
    }

    fmt.Printf("%v\n", r)



    fmt.Println("main exit")
}