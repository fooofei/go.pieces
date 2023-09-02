package main

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"os"
	"strings"
)

type Chain struct {
	From         *Url
	To           *Url
	KeyFilePath  string
	CertFilePath string
	HttpErrLog   *log.Logger
}

// parseChain 解析配置文件
func parseChainList(filePath string) ([]Chain, error) {
	var content, err = os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	var s = bufio.NewScanner(bytes.NewReader(content))
	var pairs = make([]Chain, 0)
	for s.Scan() {
		var line = s.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}
		var pair = strings.Fields(line)
		if len(pair) < 2 {
			continue
		}
		var s1 *Url
		var s2 *Url
		if s1, err = parsePathToUrl(pair[0]); err != nil {
			return nil, err
		}
		if s2, err = parsePathToUrl(pair[1]); err != nil {
			return nil, err
		}
		pairs = append(pairs, Chain{
			From: s1,
			To:   s2,
		})
	}
	return pairs, nil
}

func setChainListKeyLogWriter(chainList []Chain, clientKeyLogWriter, upstreamKeyLogWriter io.Writer) []Chain {
	var r = make([]Chain, 0, len(chainList))
	for _, e := range chainList {
		e.From.KeyLogWriter = clientKeyLogWriter
		e.To.KeyLogWriter = upstreamKeyLogWriter
		r = append(r, e)
	}
	return r
}

func setChainListTLSKeyCert(chainList []Chain, keyFilePath, certFilePath string) []Chain {
	var r = make([]Chain, 0, len(chainList))
	for _, e := range chainList {
		e.KeyFilePath = keyFilePath
		e.CertFilePath = certFilePath
		r = append(r, e)
	}
	return r
}

func setChainListHttpErrLog(chainList []Chain, w io.Writer) []Chain {
	var r = make([]Chain, 0, len(chainList))
	var l = log.New(w, "net/http/server ", log.Lshortfile|log.LstdFlags)
	for _, e := range chainList {
		e.HttpErrLog = l
		r = append(r, e)
	}
	return r
}

func hasHttpsChain(chainList []Chain) bool {
	for _, v := range chainList {
		if v.To.Scheme == HttpsScheme || v.From.Scheme == HttpsScheme {
			return true
		}
	}
	return false
}
