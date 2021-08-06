package test

import (
	"encoding/json"
	"gotest.tools/assert"
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
