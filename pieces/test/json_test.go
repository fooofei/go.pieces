package test

import (
	"encoding/json"
	"gotest.tools/v3/assert"
	"testing"
)

type level2Struct struct {
	Level2 string `json:"level_2"`
}

type toMarshal struct {
	Level1 string        `json:"level_1"`
	Level2 *level2Struct `json:"level_2"`
}

func TestMarshalNil(t *testing.T) {
	body := []byte(`
{
 "level_1": "data1"
}
`)
	m := &toMarshal{}
	err := json.Unmarshal(body, &m)
	if err != nil {
		t.Error(err)
		return
	}
	// json 里没有写 level_2 那么这个成员就是 nil
	assert.Equal(t, m.Level2 == nil, true)
}

func TestMarshalEmpty(t *testing.T) {
	body := []byte(`
{
  "level_1": "data1",
  "level_2": ""
}
`)
	m := &toMarshal{}
	err := json.Unmarshal(body, &m)
	// err = {error | *encoding/json.UnmarshalTypeError}
	// Value = {string} "string"
	// Type = {reflect.Type | *reflect.rtype}
	// Offset = {int64} 40
	// Struct = {string} "toMarshal"
	// Field = {string} "level_2"
	assert.Equal(t, err != nil, true)
	assert.Equal(t, m.Level2.Level2, "")
}


// 后续研究 如何保持 json 带注释
// 后续研究 如何保持 json 序列化的输出与反序列化前一致

// "github.com/buger/jsonparser" 可以保持原有的结构 进行针对性替换