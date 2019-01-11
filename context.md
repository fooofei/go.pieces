
> context in go

    golang 中处理 goroutine 的并发控制，存在 3 个，
    sync.WaitGroup/channel/Context.
    
    Context 我还没用过，觉得这个很难理解。
    
    这里一个文章分析了 Context 不应该存在的原因。
    https://faiface.github.io/post/context-should-go-away-go2/
    
    有些文章说 Context 存在的好处是让 sub routine 树形 cancel。
    但是我现在都是 channel 来做到这个效果的。