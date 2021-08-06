package test

import (
	"gotest.tools/assert"
	"testing"
)

func TestGetFuncNameFromAnotherFile(t *testing.T) {

	name1 := GetFuncName(TestGetFuncNameFromAnotherFile)
	assert.Equal(t, name1, "TestGetFuncNameFromAnotherFile")

}
