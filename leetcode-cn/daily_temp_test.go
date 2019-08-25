package leetcode

import (
	"container/list"
	"gotest.tools/assert"
	"sort"
	"testing"
)

// https://leetcode-cn.com/problems/daily-temperatures/
type withIdx struct {
	value int
	index int
}

func dailyTemperaturesBad(T []int) []int {
	result := make([]int, len(T))
	sortedSubSrc := make([]withIdx, 0)
	idx := len(T)
	for idx -= 1; idx >= 0; idx-- {
		l := len(sortedSubSrc)
		value := T[idx]
		// binary search
		find := sort.Search(l, func(i int) bool {
			return sortedSubSrc[i].value > value
		})
		newWithIdx := withIdx{value: value, index: idx}
		if find < l {
			result[idx] = sortedSubSrc[find].index - idx
			// insert
			tmpSlice := append(sortedSubSrc[0:find], newWithIdx)
			sortedSubSrc = append(tmpSlice, sortedSubSrc[find:]...)
		} else {
			result[idx] = 0
			// append to end
			sortedSubSrc = append(sortedSubSrc, newWithIdx)
		}
	}
	return result
}

func dailyTemperaturesBack(T []int) []int {
	result := make([]int, len(T))
	idx := len(T)
	for idx -= 2; idx >= 0; idx-- {
		if T[idx] < T[idx+1] {
			result[idx] = 1
		} else {
			for idxDummy := idx + 1 + result[idx+1];
				idxDummy < len(T);
			idxDummy += result[idxDummy] {
				if T[idxDummy] > T[idx] {
					result[idx] = idxDummy - idx
					break
				}
				if result[idxDummy] == 0 {
					break
				}
			}
		}
	}
	return result
}

func dailyTemperatures(T []int) []int {
	result := make([]int, len(T))
	stack := list.New()

	for idx, value := range T {
		for stack.Back() != nil {
			back := stack.Back().Value.(*withIdx)
			if value > back.value {
				result[back.index] = idx - back.index
				stack.Remove(stack.Back())
			} else {
				break
			}
		}
		stack.PushBack(&withIdx{value: value, index: idx})
	}
	return result
}

func TestDailyTemp(t *testing.T) {
	input := []int{73, 74, 75, 71, 69, 72, 76, 73}
	expect := []int{1, 1, 4, 2, 1, 1, 0, 0}

	result := dailyTemperatures(input)
	t.Logf("expect = %v", expect)
	t.Logf("resut= %v", result)

	assert.DeepEqual(t, result, expect)

}
