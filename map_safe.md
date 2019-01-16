

### 共享读 map，不写 可以

https://stackoverflow.com/questions/11063473/map-with-concurrent-access

### RWLock&map  vs sync.Map

https://github.com/deckarep/sync-map-analysis

https://medium.com/@deckarep/the-new-kid-in-town-gos-sync-map-de24a6bf7c2c


### 推荐map使用技巧

不吝啬使用chan，把对 map 对操作聚拢到1个 go routine 里，在1个 go routine 里

对map操作，协程安全。 