
package yaml

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/vmware-labs/yaml-jsonpath/pkg/yamlpath"
	old "gopkg.in/yaml.v3" // 以后可以尝试使用 github.com/goccy/go-yaml
)

// Marshal 把输入的 in 对象序列化为字节流
// 使用熟悉的 2 空格缩进
func Marshal(in any) ([]byte, error) {
	var b = bytes.NewBufferString("")
	var yamlEncoder = old.NewEncoder(b)
	yamlEncoder.SetIndent(2)
	if err := yamlEncoder.Encode(in); err != nil {
		return nil, fmt.Errorf("failed marshal yaml, %w", err)
	}
	return b.Bytes(), nil
}

func MarshalWriter(in any, w io.Writer) error {
	var b, err = Marshal(in)
	if err != nil {
		return err
	}
	_, err = w.Write(b)
	return err
}

// UnmarshalTo 把 in 对象序列化后，然后反序列化生成 out，用于不同对象之间的转换
func UnmarshalTo(in any, out any) error {
	var content, err = old.Marshal(in)
	if err != nil {
		return fmt.Errorf("failed yaml marshal, %w", err)
	}
	return old.Unmarshal(content, out)
}

func UnmarshalReader(r io.Reader, out any) error {
	var content, err = io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("failed read, %w", err)
	}
	return UnmarshalBytes(content, out)
}

func UnmarshalBytes(content []byte, out any) error {
	return old.Unmarshal(content, out)
}

// UnmarshalToDocList 解析存在多个 yaml 文件拼接的字节流，生成 list
func UnmarshalToDocList(content []byte) ([]*old.Node, error) {
	var docList = make([]*old.Node, 0)
	var dec = old.NewDecoder(bytes.NewReader(content))
	for {
		var value = &old.Node{} // 反序列化为  node 类型可以保留yaml文件中的 comment
		// pass a reference to spec reference
		var err = dec.Decode(value)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed parse yaml, %w", err)
		}
		docList = append(docList, value)
	}
	return docList, nil
}

func GetNodeString(node *old.Node) (string, bool) {
	if node.Tag == "!!str" {
		return node.Value, true
	}
	return "", false
}

func SetNodeString(node *old.Node, value string) {
	node.SetString(value)
}

func SetNodeValue(node *old.Node, value any) error {
	switch v := value.(type) {
	case string:
		SetNodeString(node, v)
		return nil
	case int:
		node.Kind = old.ScalarNode
		node.Tag = "!!int"
		node.Value = strconv.Itoa(v)
		return nil
	case *old.Node:
		node.Kind = old.MappingNode
		node.Tag = node.ShortTag() // 提供 kind 后就可以自动获取 tag
		node.Content = append(node.Content, v)
		return nil
	default:
		return fmt.Errorf("unsupport type %T %v", value, value)
	}
}

func NewStringNode(value string) *old.Node {
	var n = &old.Node{}
	n.SetString(value)
	return n
}

func NewIntNode(value int) *old.Node {
	var n = &old.Node{
		Kind:  old.ScalarNode,
		Tag:   "!!int",
		Value: strconv.Itoa(value),
	}
	return n
}

func NewBytesNode(value []byte) *old.Node {
	var n = &old.Node{}
	n.SetString(string(value))
	return n
}

// FindNode 找到1个可用的 node，是开源库的瘦身版本
func FindNode(node *old.Node, pathList []string) (*old.Node, error) {
	const pathSep = "." // 开源库约定是这个分割符号
	var path = strings.Join(pathList, pathSep)
	return FindNodeRawPath(node, fmt.Sprintf("$.%v", path))
}

func FindNodeRawPath(node *old.Node, path string) (*old.Node, error) {
	var pathFinder, err = yamlpath.NewPath(path)
	if err != nil {
		return nil, fmt.Errorf("failed new path finder, %w", err)
	}
	var resultList []*old.Node
	resultList, err = pathFinder.Find(node)
	if err != nil {
		return nil, fmt.Errorf("failed yaml find, %w", err)
	}
	if len(resultList) < 1 {
		return nil, fmt.Errorf("not exists path %v", path)
	}
	return resultList[0], nil
}

// 我用开源软件的实现代替这里了
// 根据 kind（例如为 Deployment）和 path (a#b#c) 来获取是否存在值
func yamlGetDeprecated(node *old.Node, kind string, pathList []string) (any, error) {
	if len(pathList) < 1 {
		return nil, fmt.Errorf("not given path")
	}
	var root = make(map[string]any)
	if err := UnmarshalTo(node, &root); err != nil {
		return nil, fmt.Errorf("failed convert node to map, %w", err)
	}
	if root["kind"] != kind {
		return nil, fmt.Errorf("mismatch kind, %v != %v", root["kind"], kind)
	}

	for i, pathKey := range pathList {
		var va, ok = root[pathKey]
		if !ok {
			return nil, fmt.Errorf("not exits %v", strings.Join(pathList[:i+1], ","))
		}
		if i == len(pathList)-1 {
			return va, nil
		}
		switch v := va.(type) {
		case map[string]any:
			root = v
		case map[string]string:
			// root = mapStringToAny(v)
		default:
			root = make(map[string]any)
		}
	}

	return nil, fmt.Errorf("not exists")
}
