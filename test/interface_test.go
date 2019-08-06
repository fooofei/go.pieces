package go_pieces

import (
	"fmt"
	"gotest.tools/assert"
	"testing"
)

func switchInterface(inInterface interface{}) (interface{}, error) {
	switch inInterface.(type) {
	case string:
		return "a string", nil
	case int:
		return 100, nil
	default:
		return nil, fmt.Errorf("unknown type %T", inInterface)
	}
}

func TestSwitchInterface(t *testing.T) {
	v1, v2 := switchInterface("ok")
	assert.Equal(t, v1, "a string")
	assert.NilError(t, v2)

	v1, v2 = switchInterface(0)
	assert.Equal(t, v1, 100)
	assert.NilError(t, v2)

	v1, v2 = switchInterface(false)
	assert.Equal(t, v1, nil)
	assert.ErrorContains(t, v2, "unknown")
	t.Log(v2)
}

func oneByOneInterface(inInterface interface{}) (interface{}, error) {
	if _, is := inInterface.(string); is {
		return "a string", nil
	}
	if _, is := inInterface.(int); is {
		return 100, nil
	}
	return nil, fmt.Errorf("unknown type %T", inInterface)
}

func TestOneByOneInterface(t *testing.T) {
	v1, v2 := oneByOneInterface("ok")
	assert.Equal(t, v1, "a string")
	assert.NilError(t, v2)

	v1, v2 = oneByOneInterface(0)
	assert.Equal(t, v1, 100)
	assert.NilError(t, v2)

	v1, v2 = oneByOneInterface(false)
	assert.Equal(t, v1, nil)
	assert.ErrorContains(t, v2, "unknown")
	t.Log(v2)
}

func TestTypeCastCheck(t *testing.T) {

	a := "a string"

	// _ = a.(string) //error  (non-interface type string on left)
	s1 := interface{}(a).(string)
	t.Logf(s1)
	assert.Equal(t, s1, a)

	s2, is := interface{}(a).(string)
	assert.Equal(t, s2, a)
	assert.Equal(t, is, true)
}
