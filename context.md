
> context in go

    golang 中处理 goroutine 的并发控制，存在 3 个，
    sync.WaitGroup/channel/Context.
    
    Context 我还没用过，觉得这个很难理解。
    
    这里一个文章分析了 Context 不应该存在的原因。
    https://faiface.github.io/post/context-should-go-away-go2/
    
    有些文章说 Context 存在的好处是让 sub routine 树形 cancel。
    但是我现在都是 channel 来做到这个效果的。
    
    用过了 效果还可以。
    以前我用 chan 做关闭routine 通知。
    后来因为 net.Conn 库自己用 context.Context ，索性我就跟他统一了。
    现在的使用组合是，sync.WaitGroup 用来在main routine 等待 sub routine 
    全部结束， main routine 获取 CTRL +C 通过 context.Context来通知
    sub routine 提前结束任务。