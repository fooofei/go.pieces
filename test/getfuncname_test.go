package go_pieces

import (
	"gotest.tools/assert"
	"path"
	"reflect"
	"runtime"
	"testing"
)

func canYouTellMe() {

}

func TestGetFuncName(t *testing.T) {
	name1 := GetFuncName(canYouTellMe)
	name2 := GetFuncName(TestGetFuncName)

	assert.Equal(t, name1, "canYouTellMe")

	assert.Equal(t, name2, "TestGetFuncName")
}

func GetFuncName(f interface{}) string {
	a := reflect.ValueOf(f)
	b := a.Pointer()
	c := runtime.FuncForPC(b)
	// github.com/fooofei/go_pieces/test.canYouTellMe
	fullPath := c.Name()
	// test.canYouTellMe
	moduleName := path.Base(fullPath)
	ext := path.Ext(moduleName)
	return ext[1:]
}
