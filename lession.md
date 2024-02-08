
## 新手入门	
微软的 Go 语言入门教程 https://docs.microsoft.com/zh-cn/learn/paths/go-first-steps/

[the-way-to-go_ZH_CN] Go 入门指南 https://github.com/unknwon/the-way-to-go_ZH_CN

[Go语言核心36讲]作者郝林http://ilearning.huawei.com/edx/next/courses-view?courseId=100013101

从 C 语言转到 Go 语言 https://hyperpolyglot.org/c

[从代码片段学习 Go 语言语法]Go By Assertion https://garba.org/article/general/go-by-assertion/go-by-assertion.html#

[速查表][Go Cheat Sheet] https://github.com/a8m/golang-cheat-sheet

[另一个学习教程] https://golangr.com/

[Golang修养之路] https://github.com/aceld/golang

[深入Go Module之go.mod文件解析] https://colobu.com/2021/06/28/dive-into-go-module-1/

[深入Go Module之讨厌的v2] https://colobu.com/2021/06/28/dive-into-go-module-2/

## 为 Java 转语言准备
从Java到Golang快速入门 https://www.flysnow.org/2016/12/28/from-java-to-golang.html

From Java to Go https://gquintana.github.io/2017/01/15/From-Java-to-Go.html

Lessons learned porting 50k loc from Java to Go https://blog.kowalczyk.info/article/19f2fe97f06a47c3b1f118fd06851fad/lessons-learned-porting-50k-loc-from-java-to-go.html

Java to Go in-depth tutorial https://yourbasic.org/golang/go-java-tutorial/

## 高手进阶	
[Go 语言设计与实现] https://draveness.me/golang/

[coolshell]  GO编程模式 https://coolshell.cn/articles/21228.html

[Tony Bai ] Gopher Daily https://gopher-daily.com/issues/202104/issue-20210413.md

[gocn_news_set] 爬取go中国技术社区... https://github.com/Han-Ya-Jun/gocn_news_set

[Golang 性能诊断] https://mp.weixin.qq.com/s/TjOxsotZ68XpKQviNlqQmQ

[8 Essential Go Module tidbits for a newbie] https://zaracooper.github.io/blog/posts/go-module-tidbits/

[通过实例理解Go Execution Tracer] https://mp.weixin.qq.com/s/L7HiqA02g-l-b2pD6aRtbw

## Go 语言与容器	 
基于 Go 语言构建的既小又安全的微服务容器
https://chemidy.medium.com/create-the-smallest-and-secured-golang-docker-image-based-on-scratch-4752223b7324

Go 语言与容器 GOMAXPROCS
http://www.dockone.io/article/9387

## 陷阱	
pprof 暴露了你的数据 https://mmcloughlin.com/posts/your-pprof-is-showing

不要在生产环境使用 http.DefaultServerMux？ https://mp.weixin.qq.com/s/ldMdmrC32f0o8pQvpPzvCw

[go build 是否真的是静态编译了？]https://stackoverflow.com/questions/36279253/go-compiled-binary-wont-run-in-an-alpine-docker-container-on-ubuntu-host

[使用 Go 内建 net.HTTP 使用的一些陷阱] Gotchas in the Go Network Packages Defaults https://martin.baillie.id/wrote/gotchas-in-the-go-network-packages-defaults/


## 优雅编码	

推荐使用函数来构造struct https://web3.coach/golang-why-you-should-use-constructors

[Go Antipatterns] https://hackmysql.com/golang/go-antipatterns/

## context.Context 应该作为函数的参数吗

[Should we pass go context.Context to all the components? #2716](https://github.com/dapr/dapr/issues/2716)

Passing context.Context to all relevant methods, golang https://stackoverflow.com/questions/61707002/passing-context-context-to-all-relevant-methods-golang

Do not store Contexts inside a struct type; instead, pass a Context explicitly to each function that needs it.
https://pkg.go.dev/context

## Error

不应该使用 github.com/pkg/errors 包了，built-in errors 包够用了

推荐定义一种新的 error 对象可以获取 status code https://boldlygo.tech/posts/2024-01-08-error-handling/
