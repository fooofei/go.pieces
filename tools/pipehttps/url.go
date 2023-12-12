package main

import (
	"fmt"
	"io"
	"net"
	"net/url"
	"strings"
)

const (
	HttpsScheme = "https"
	HttpScheme  = "http"
)

type Url struct {
	Scheme       string
	Host         string
	Port         string
	KeyLogWriter io.Writer
}

func (s Url) URL() string {
	return fmt.Sprintf("%v://%v", s.Scheme, s.ListenAddr())
}

func (s Url) ListenAddr() string {
	if s.Port == "" {
		return s.Host
	}
	return net.JoinHostPort(s.Host, s.Port)
}

// 解析这样的格式 https://0.0.0.0:443 字符串到 url 对象
func parsePathToUrl(path string) (Url, error) {
	var s = Url{}
	var u, err = url.Parse(path)
	if err != nil {
		return Url{}, fmt.Errorf("failed url.Parse '%v' %w", path, err)
	}
	s.Scheme = u.Scheme
	s.Scheme = strings.ToLower(s.Scheme)
	var host, port string
	if host, port, err = net.SplitHostPort(u.Host); err == nil {
		s.Host = host
		s.Port = port
	} else {
		s.Host = u.Host
		s.Port = ""
	}

	if s.Port == "" {
		if s.Scheme == HttpsScheme {
			s.Port = "443"
		} else {
			s.Port = "80"
		}
	}
	return s, nil
}

func ParserUrlList(pathList []string) ([]Url, error) {
	var resultList = make([]Url, 0)

	for _, p := range pathList {
		var r, err = parsePathToUrl(p)
		if err != nil {
			return nil, err
		}
		resultList = append(resultList, r)
	}
	return resultList, nil
}

func hasHttpsUrl(urlList []Url) bool {
	for _, u := range urlList {
		if u.Scheme == HttpsScheme {
			return true
		}
	}
	return false
}
