package main

import "sort"

type sortItem struct {
	Result   string
	ItemList []*requestItem
}

type sortItemList []*sortItem

func (p sortItemList) Len() int               { return len(p) }
func (p sortItemList) Less(i int, j int) bool { return len(p[i].ItemList) < len(p[j].ItemList) }
func (p sortItemList) Swap(i int, j int)      { p[i], p[j] = p[j], p[i] }

func sortResultList(itemList []*requestItem) []*sortItem {
	var list = make([]*sortItem, 0)
	var m = make(map[string]*sortItem, 0)
	for _, item := range itemList {
		m[item.Result] = &sortItem{
			Result:   item.Result,
			ItemList: make([]*requestItem, 0),
		}
	}
	for _, item := range itemList {
		m[item.Result].ItemList = append(m[item.Result].ItemList, item)
	}
	for _, v := range m {
		list = append(list, v)
	}
	sort.Sort(sortItemList(list))
	return list
}
