
试用总结


## YAP in Go 

用法就是 yap 写了一个 go package，这是一个简易的 wrapper， 可以简单 c.TEXT c.JSON 方法返回数据

写的代码还是 go 语言

## YAP classfile v1 

用法就是写一个文本的 main.yap 文件，这是一个文本语言

只写简单的 go 语言代码，可以 import go 的库，可以使用 go 的函数

`${xx}` 被解释为调用 Context.Gop_Env(xx) 这个是从 内建的 request.FormValue 获取值

## YAP classfile v2

比 v1 的改进点是可以把不同的 path 写成文件名称的形式，这样不用写在一个大文件中