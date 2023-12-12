package main

import (
	"bufio"
	"bytes"
	"os"
	"strings"
)

// ParseDnsMapper 解析dns 映射文件
// 文件内容为
// mapper file format:
// 127.0.0.1:18100 example.com:1984
// this will connect to 127.0.0.1:18100 when receive header Host: example.com:1984
// 尽量与 /etc/hosts 格式一致
func ParseDnsMapper(filePath string) (map[string]string, error) {
	var content, err = os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	var s = bufio.NewScanner(bytes.NewReader(content))
	var result = make(map[string]string, 0)
	for s.Scan() {
		var line = s.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}
		var pair = strings.Fields(line)
		if len(pair) < 2 {
			continue
		}
		result[pair[1]] = pair[0]
	}
	return result, nil
}
