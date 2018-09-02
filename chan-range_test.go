package main

import (
	"fmt"
)


func produceForRange(msgQ chan<- string){
    for i:=0;i<=5;i+=1{
        msgQ <- fmt.Sprintf("push msg %v", i)
    }
    close(msgQ) // 阻塞在这个chan的等待都会守信
}

func ExampleChanRange(){
    var msgQ = make(chan string)
    go produceForRange(msgQ)

    for msg := range msgQ{
        fmt.Printf("Main get from msgQ:%v\n",msg)
    }
    fmt.Printf("%v exit\n",GetFuncName(ExampleChanRange))
    // output:
    // Main get from msgQ:push msg 0
    //Main get from msgQ:push msg 1
    //Main get from msgQ:push msg 2
    //Main get from msgQ:push msg 3
    //Main get from msgQ:push msg 4
    //Main get from msgQ:push msg 5
    //ExampleChanRange exit
}
