
2024/09/04 实际测试， 

pkg/a
  a.go
  a_test.go 文件中写了 
```go
func init() {

}
```
那么在当前文件用例执行前会执行，但是
如果另外的文件是 import pkg/a  使用 a 包内的函数，则 init 函数不会执行
