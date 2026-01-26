package test


// 生成轮询endpoints列表， 输入 lastList 为上一轮输出的结果列表，输入 newList 为即时最新的实例列表，经过了排序，
// 这两个列表可能个数不相等
func roundrobinEndpoints(lastList []string, newList []string) []string {
	var lastMap = slicesToMap(lastList)
	var newMap = slicesToMap(newList)

	var returnList = slices.Grow(lastList, len(newList)+1) // 增长空间，最坏容纳下新列表，+1是因为尾部插入一个

	returnList = lastList

	// 把在 newList 中的不在 lastList 中的填充进去
	returnList = slicesAppendFunc(returnList, newList, func(s string) bool {
		_, exist := lastMap[s]
		return !exist
	})
	// 保留此顺序，把第一个放到最后一个
	if len(lastList) > 1 {
		returnList = append(returnList, lastList[0])
		returnList = returnList[1:]
	}
	// 删除在 newList 中没有的
	returnList = slices.DeleteFunc(returnList, func(s string) bool {
		_, exist := newMap[s]
		return !exist
	})
	return returnList
}

func slicesToMap(slice []string) map[string]struct{} {
	var m = make(map[string]struct{})
	for _, i := range slice {
		m[i] = struct{}{}
	}
	return m
}

func slicesAppendFunc(dstList, srcList []string, fn func(string) bool) []string {
	for _, i := range srcList {
		if fn(i) {
			dstList = append(dstList, i)
		}
	}
	return dstList
}
