package test 


// 
// slices.Clip  回收未使用的 cap 空间 
// slices.Grow 增加 cap 空间
// slices.Clone 浅拷贝 
// slices.Compact
//     压缩连续相同的元素，{0, 1, 1, 2, 3, 5, 8} 压缩为 [0 1 2 3 5 8] 
//       {"1", "1", "2", "1"} 压缩为 [1 2 1]
