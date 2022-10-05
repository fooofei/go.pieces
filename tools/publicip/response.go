package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// ResponseParser defines get ip from url response text
type ResponseParser func(r io.Reader) (string, error)

// http://ifcfg.cn/echo
func getIpInJsonIP(r io.Reader) (string, error) {
	var b, err = io.ReadAll(r)
	if err != nil {
		return "", err
	}
	var m = make(map[string]interface{})
	_ = json.Unmarshal(b, &m)
	var ip, ok = m["ip"].(string)
	if ok {
		return ip, nil
	}
	return "", fmt.Errorf("not found ip in %v", string(b))
}

// GetIpInJsonIpIp parse result for https://ip.nf/me.json
func getIpInJsonIPIP(r io.Reader) (string, error) {
	var b, err = io.ReadAll(r)
	if err != nil {
		return "", err
	}
	var m = make(map[string]interface{})
	_ = json.Unmarshal(b, &m)
	var ipObj, ok = m["ip"].(map[string]interface{})
	if ok {
		var ip, ok = ipObj["ip"].(string)
		if ok {
			return ip, nil
		}
	}
	return "", fmt.Errorf("not found ip in %v", string(b))
}

// GetIpInJsonQuery parse result for http://ip-api.com/json
func getIpInJsonQuery(r io.Reader) (string, error) {
	var b, err = io.ReadAll(r)
	if err != nil {
		return "", err
	}
	var m = make(map[string]interface{})
	_ = json.Unmarshal(b, &m)
	var ip, ok = m["query"].(string)
	if ok {
		return ip, nil
	}
	return "", fmt.Errorf("not found ip in %v", string(b))
}

// GetIpInJsonYourFuck parse result for https://wtfismyip.com/json
func getIpInJsonYourFuck(r io.Reader) (string, error) {
	var b, err = io.ReadAll(r)
	if err != nil {
		return "", err
	}
	var m = make(map[string]interface{})
	_ = json.Unmarshal(b, &m)
	var ip, ok = m["YourFuckingIPAddress"].(string)
	if ok {
		return ip, nil
	}
	return "", fmt.Errorf("not found ip in %v", string(b))
}

// GetIpInPlainText return ip from plain text
// https://api.ipify.org
// https://ip.seeip.org
// https://ifconfig.me/ip
// https://ifconfig.co/ip
func getIpInPlainText(r io.Reader) (string, error) {
	var b, err = io.ReadAll(r)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}

// http://ip.taobao.com/service/getIpInfo2.php?ip=myip
func getIpInJsonTaobao(r io.Reader) (string, error) {
	var b, err = io.ReadAll(r)
	if err != nil {
		return "", err
	}
	var m = make(map[string]interface{})
	_ = json.Unmarshal(b, &m)
	if data, ok := m["data"].(map[string]interface{}); ok {
		if ip, ok := data["ip"].(string); ok {
			return ip, nil
		}
	}
	return "", fmt.Errorf("not found ip in %v", string(b))
}
