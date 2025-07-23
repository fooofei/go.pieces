package test

import (
	"bytes"
	"fmt"
	"text/template"
)

// 把 ${} 分割的变量渲染为最终值 不支持 ${var:defaultvalue} 的语法
func renderActions(text string, values map[string]any) (string, error) {
	// custom varable delim `${}`
	var leftDelim = "${"
	var rightDelim = "}"

	// 这个分割符号是 action 的分割符号，因此要把 values 作为 funcs 传入
	var tmpl = template.New("test").Delims(leftDelim, rightDelim).Funcs(valuesToFuncs(values))
	var err error

	// 调试action是否定义 text/template/parse/parse.go:764
	if tmpl, err = tmpl.Parse(text); err != nil {
		return "", fmt.Errorf("failed template parse, %w", err)
	}

	var b = bytes.NewBufferString("")
	if err = tmpl.Execute(b, values); err != nil {
		return "", fmt.Errorf("failed template execute, %w", err)
	}
	return b.String(), nil
}

func valuesToFuncs(values map[string]any) map[string]any {
	var r = make(map[string]any)
	for k, v := range values {
		r[k] = func() any {
			return v
		}
	}
	return r
}
