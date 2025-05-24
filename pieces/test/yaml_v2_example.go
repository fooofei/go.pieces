package yaml

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/goccy/go-yaml"
	yast "github.com/goccy/go-yaml/ast"
)

// yaml.YAMLToJSON 的原理是会设置为 flowStyle 然后就会调用到 printer.PrintNode %v 打印就会调用到 node 的 String() 方法
//func (n *MappingNode) flowStyleString(commentMode bool) string {
//这里面是手工拼接的 {} 号

func Marshal(in any, cm yaml.CommentMap) ([]byte, error) {
	return yaml.MarshalWithOptions(in, yaml.Indent(2), yaml.UseLiteralStyleIfMultiline(true),
		yaml.WithComment(cm))
}

func UnmarshalBytes(content []byte, out any) error {
	return yaml.UnmarshalWithOptions(content, out)
}

// UnmarshalBytesWithComment 反序列化为 node 和 注释，这两个只能分开，没有合并的办法
func UnmarshalBytesWithComment(content []byte, out any) (yaml.CommentMap, error) {
	var cm = yaml.CommentMap{}
	var dec = yaml.NewDecoder(bytes.NewBuffer(content), yaml.CommentToMap(cm))
	if err := dec.DecodeContext(context.Background(), out); err != nil {
		return nil, fmt.Errorf("failed decode yaml. %w", err)
	}
	return cm, nil
}

func UnmarshalToDocListv2(content []byte) ([]*yast.DocumentNode, error) {
	var doc = &yast.File{}
	var err = UnmarshalBytes(content, doc)
	if err != nil {
		return nil, fmt.Errorf("failed read file, %w", err)
	}
	return doc.Docs, nil
}

func MarshalToWriter(in any, w io.Writer) error {
	var cm yaml.CommentMap
	var b, err = Marshal(in, cm)
	if err != nil {
		return err
	}
	_, err = w.Write(b)
	return err
}

func MarshalNodeToWriter(in yast.Node, cm yaml.CommentMap, w io.Writer) error {
	//yaml.Marshal() // 不要使用库的导出函数Marshal 会丢失注释  // EncodeContext 其实就是调用的 printer.Printer.PrintNode
	//var b, err = in.MarshalYAML() // 不要使用node的导出函数Marshal 会丢失注释 调用的是 n.String()
	//   Formatter 最好，但是这个包不导出
	var b, err = Marshal(in, cm)
	if err != nil {
		return fmt.Errorf("failed marshal yaml, %w", err)
	}
	//var b = FormatNode(in)
	_, err = w.Write([]byte(b))
	return err
}

// UnmarshalTo 把 in 对象序列化后，然后反序列化生成 out，用于不同对象之间的转换
func UnmarshalTo(in any, out any) error {
	var content, err = yaml.Marshal(in)
	if err != nil {
		return fmt.Errorf("failed yaml marshal, %w", err)
	}
	return yaml.Unmarshal(content, out)
}

// FindNode 找到1个可用的 node，是开源库的瘦身版本
func FindNode(node yast.Node, pathList []string) (yast.Node, error) {
	const pathSep = "." // 开源库约定是这个分割符号
	var path = strings.Join(pathList, pathSep)
	var p, err = yaml.PathString(fmt.Sprintf("$.%v", path))
	if err != nil {
		return nil, fmt.Errorf("failed make yaml PathString, %w", err)
	}
	return p.FilterNode(node)
}
