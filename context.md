
## `context.Contenxt` in golang

golang 中处理 goroutine 的并发控制，存在 3 个技术

`sync.WaitGroup`/ `channel`/ `context.Context`.
    

这里一个文章 [context-should-go-away-go2](https://faiface.github.io/post/context-should-go-away-go2/) 分析了 `Context` 不应该存在的原因。


当 goroutine 规模小的时候，还可以以来 chan 来控制并发，上量之后，就需要 context 上场了。
    
Context 存在的好处是让 sub routine 树形 cancel。

使用技法是：sync.WaitGroup 用来在 main routine 等待 sub routine 全部结束， main routine 捕获 CTRL +C signal, 通过 context.Context 来通知 sub routine 提前退出。
