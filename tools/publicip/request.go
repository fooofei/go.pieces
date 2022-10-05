package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

type requestItem struct {
	Uri       string
	ParseFunc ResponseParser
	Result    string
	TakeTime  time.Duration
	Err       error
}

// request will request to uri for get ip from response body
func request(ctx context.Context, uri string, parseFunc ResponseParser) (string, error) {
	var clt = http.DefaultClient
	var req, err = http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return "", err
	}
	var resp *http.Response
	if resp, err = clt.Do(req); err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("invalid response status %v", resp.Status)
	}
	var result string
	if result, err = parseFunc(resp.Body); err != nil {
		return "", err
	}
	return strings.TrimSpace(result), nil
}

func sendAll(ctx context.Context, itemList []*requestItem) {
	var wg = &sync.WaitGroup{}
	for _, item := range itemList {
		wg.Add(1)
		go func(whichItem *requestItem) {
			var start = time.Now()
			whichItem.Result, whichItem.Err = request(ctx, whichItem.Uri, whichItem.ParseFunc)
			whichItem.TakeTime = time.Since(start)
			wg.Done()
		}(item)
	}
	wg.Wait()
}
