
2013-02-16

Concurrency is not parallelism 并发和并行的不同概念

那句英文解释起来挺无力的

`Concurrency is about dealing with logs of things t once.`

`parallelism is about doing lots of things at once.`


2010-09-23

concurrency timeout

一个 channel 用来存放超时
```go
timeout := make(chan bool,1)
go func(){
    time.Slee(1 * time.Second)
    timeout < true
}()
```
这个例子可以用 `time.After`来代替，很简洁

它可以这样搭配使用

```go
select {
    case <- ch:
        //一个用来读业务数据的 chan
    case <- timeout:
        //超时了
}
```


2010-07-13

Share Memory By Communication

`Do not communicate by sharing memory; instead, share memory by communicating.`


