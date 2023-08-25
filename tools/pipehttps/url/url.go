package url

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
)

type Url struct {
	Scheme string
	Host   string
	Port   string
}

func (s *Url) URL() string {
	return fmt.Sprintf("%v://%v", s.Scheme, s.Join())
}

func (s *Url) Join() string {
	if s.Port == "" {
		return s.Host
	}
	return net.JoinHostPort(s.Host, s.Port)
}

type Chain struct {
	From *Url
	To   *Url
}

func Parse(path string) (*Url, error) {
	var s = &Url{}
	var u, err = url.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("failed url.Parse '%v' %w", path, err)
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
		if s.Scheme == "https" {
			s.Port = "443"
		} else {
			s.Port = "80"
		}
	}
	return s, nil
}

// ParseChain 解析配置文件
func ParseChain(filePath string) ([]Chain, error) {
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
			return nil, fmt.Errorf("invalid format of '%s'", line)
		}
		var s1 *Url
		var s2 *Url
		if s1, err = Parse(pair[0]); err != nil {
			return nil, err
		}
		if s2, err = Parse(pair[1]); err != nil {
			return nil, err
		}
		pairs = append(pairs, Chain{
			From: s1,
			To:   s2,
		})
	}
	return pairs, nil
}
