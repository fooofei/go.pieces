package main

import (
    "fmt"
    "log"
    "os"
    "time"
)


func routine(idx int, v * byte){
    log.Printf("routine %v got %v", idx, v)
}

func main(){

    log.SetFlags(log.Lshortfile | log.LstdFlags)
    log.SetPrefix(fmt.Sprintf("pid= %v ",os.Getpid()))

    ar := make([]byte,5)

    for idx,_ := range ar {
        log.Printf("&ar[%v]=%v",idx, &ar[idx])
    }

    // wrong address
    for idx,_ := range ar {
        go func() {
            routine(idx, &ar[idx])
        }()
    }

    time.Sleep(time.Second*3)

    // right address
    for idx,_ := range ar {
        go routine(idx+10, &ar[idx])
    }

    time.Sleep(time.Second*3)

    log.Printf("main exit")

}
//pid= 37989 2019/02/26 15:29:54 goroutine.go:23: &ar[0]=0xc00006c018
//pid= 37989 2019/02/26 15:29:54 goroutine.go:23: &ar[1]=0xc00006c019
//pid= 37989 2019/02/26 15:29:54 goroutine.go:23: &ar[2]=0xc00006c01a
//pid= 37989 2019/02/26 15:29:54 goroutine.go:23: &ar[3]=0xc00006c01b
//pid= 37989 2019/02/26 15:29:54 goroutine.go:23: &ar[4]=0xc00006c01c
//pid= 37989 2019/02/26 15:29:54 goroutine.go:12: routine 4 got 0xc00006c01c
//pid= 37989 2019/02/26 15:29:54 goroutine.go:12: routine 4 got 0xc00006c01c
//pid= 37989 2019/02/26 15:29:54 goroutine.go:12: routine 4 got 0xc00006c01c
//pid= 37989 2019/02/26 15:29:54 goroutine.go:12: routine 4 got 0xc00006c01c
//pid= 37989 2019/02/26 15:29:54 goroutine.go:12: routine 4 got 0xc00006c01c
//pid= 37989 2019/02/26 15:29:57 goroutine.go:12: routine 12 got 0xc00006c01a
//pid= 37989 2019/02/26 15:29:57 goroutine.go:12: routine 14 got 0xc00006c01c
//pid= 37989 2019/02/26 15:29:57 goroutine.go:12: routine 11 got 0xc00006c019
//pid= 37989 2019/02/26 15:29:57 goroutine.go:12: routine 13 got 0xc00006c01b
//pid= 37989 2019/02/26 15:29:57 goroutine.go:12: routine 10 got 0xc00006c018