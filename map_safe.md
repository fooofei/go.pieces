

### 共享读 map，不写 可以

这是基础知识，这个帖子里有讨论：
https://stackoverflow.com/questions/11063473/map-with-concurrent-access


### map 中读的场景 > 写 

在这个场景下非常适合：
共享读 map + 读写锁


### RWLock&map  vs sync.Map

对比代码：
https://github.com/deckarep/sync-map-analysis

https://medium.com/@deckarep/the-new-kid-in-town-gos-sync-map-de24a6bf7c2c
https://blog.csdn.net/xingwangc2014/article/details/79777770

分桶+读写锁 https://github.com/orcaman/concurrent-map

### map+channel 使用方式

不吝啬使用chan，把对 map 对操作聚拢到1个 go routine 里，在1个 go routine 里

对map操作，协程安全。 

https://stackoverflow.com/questions/18192173/nice-go-idiomatic-way-of-using-a-shared-map
