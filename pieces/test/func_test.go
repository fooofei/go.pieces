package test

import (
	"gotest.tools/v3/assert"
	"testing"
)

func addr() func(int) int {
	var sum int = 0
	return func(x int) int {
		sum += x
		return sum
	}
}

func TestClosureFunc1(t *testing.T) {
	pos := addr()
	neg := addr()

	expectPos := []int{
		0, 1, 3, 6, 10,
	}

	for i := 0; i < 5; i++ {
		assert.Equal(t, pos(i), expectPos[i])
	}
	expectNeg := []int{
		0, -1, -3, -6, -10,
	}
	for i := 0; i < 5; i++ {
		assert.Equal(t, neg(-i), expectNeg[i])
	}

}

func change(v func(int) int) int {
	return v(11)
}

// closure func as param will be reference
func TestClosureFuncParam(t *testing.T) {
	var a = addr()
	assert.Equal(t, a(1), 1)
	assert.Equal(t, a(2), 3)
	assert.Equal(t, a(0), 3)
	assert.Equal(t, change(a), 14)
	assert.Equal(t, a(0), 14)

}

func TestClosureFuncCopy(t *testing.T) {
	a := addr()

	assert.Equal(t, a(1), 1)
	assert.Equal(t, a(2), 3)
	assert.Equal(t, a(0), 3)

	var b = a
	assert.Equal(t, b(0), 3)
	assert.Equal(t, b(1), 4)
	assert.Equal(t, a(0), 4) // b is ref of a
	assert.Equal(t, b(0), 4)
}
