package test

import (
	"bytes"
	"gotest.tools/v3/assert"
	"testing"

	yaml2 "gopkg.in/yaml.v2"
	yaml3 "gopkg.in/yaml.v3"
)

func TestYaml2Order(t *testing.T) {
	yamlContent := []byte(`
common:
  port: 1900
  addr: "127.0.0.1"
controller:
  port: 1231
  service: "some"
ada:
  name: "deployment"
  account: "root"
`)
	decoder := yaml2.NewDecoder(bytes.NewReader(yamlContent))
	value := make(yaml2.MapSlice, 0)
	err := decoder.Decode(&value)
	assert.NilError(t, err)
	assert.Equal(t, len(value), 3)
	assert.Equal(t, value[0].Key, "common")
	assert.Equal(t, value[2].Key, "ada")

	// 测试序列化
	// keep order https://stackoverflow.com/questions/33639269/preserving-order-of-yaml-maps-using-go
	out, err := yaml2.Marshal(value)
	assert.NilError(t, err)
	yamlContent = bytes.TrimSpace(yamlContent)
	out = bytes.TrimSpace(out)
	// 再次序列化后字符串没有双引号了，长度不一致了，但是字段顺序一致
	assert.Equal(t, len(yamlContent) == len(out), false)
}

func TestYaml3Order(t *testing.T) {
	yamlContent := []byte(`
common:
  port: 1900
  addr: "127.0.0.1"
controller:
  port: 1231
  service: "some"
ada:
  name: "deployment"
  account: "root"
`)
	decoder := yaml3.NewDecoder(bytes.NewReader(yamlContent))
	// Node 类型可以保持顺序，传递 interface{} 类型不能保持顺序
	value := &yaml3.Node{}
	err := decoder.Decode(value)
	assert.NilError(t, err)

	// 测试序列化
	outBuf := bytes.NewBufferString("")
	encoder := yaml3.NewEncoder(outBuf)
	encoder.SetIndent(2) // 与我们源内容一致 保持 2 个空格缩进
	// yaml3.Marshal 默认是 4 个空格的缩进
	err = encoder.Encode(value)
	assert.NilError(t, err)
	out := outBuf.Bytes()
	yamlContent = bytes.TrimSpace(yamlContent)
	out = bytes.TrimSpace(out)

	assert.Equal(t, len(yamlContent) == len(out), true)
	assert.DeepEqual(t, yamlContent, out)
}

func TestYamlComment(t *testing.T) {
	yamlContent := []byte(`
common:
  port: 1900
  addr: "127.0.0.1"
# a comment at here
controller:
  port: 1231
  service: "some"
ada:
  name: "deployment"
  account: "root"
`)
	decoder := yaml3.NewDecoder(bytes.NewReader(yamlContent))

	value := &yaml3.Node{}
	err := decoder.Decode(value)
	assert.NilError(t, err)


	outBuf := bytes.NewBufferString("")
	encoder := yaml3.NewEncoder(outBuf)
	encoder.SetIndent(2)

	err = encoder.Encode(value)
	assert.NilError(t, err)
	out := outBuf.Bytes()
	yamlContent = bytes.TrimSpace(yamlContent)
	out = bytes.TrimSpace(out)

	// 序列化之后 文件注释也被保持下来
	assert.Equal(t, len(yamlContent) == len(out), true)
	assert.DeepEqual(t, yamlContent, out)
}