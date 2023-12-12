package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"sync/atomic"
	"time"
)

func PipeResponse(resp *http.Response, w http.ResponseWriter) error {
	for k := range resp.Header {
		w.Header().Set(k, resp.Header.Get(k))
	}
	w.WriteHeader(resp.StatusCode)
	var _, err = io.Copy(w, resp.Body)
	return err
}

func _getServerHandleFunc(ctxg context.Context, trans *http.Transport, seq *atomic.Int64, u Url) http.HandlerFunc {
	// ServeHTTP 将会把请求 pipe 出去
	return func(w http.ResponseWriter, request *http.Request) {
		var ctx, cancel = context.WithTimeout(ctxg, time.Minute)
		var count = seq.Add(1)
		var pbuf = bytes.NewBufferString("")
		var err error
		var upstreamResp *http.Response
		defer cancel()
		defer func() {
			fmt.Print(pbuf.String())
		}()

		// 重新发送原始 request 无法填充 Host，会导致发送失败
		// failed request to upstream error: <*url.Error>Get "/path/": http: Request.RequestURI can't be set in client requests
		// 因此我们这里要重新生成 request，当前的clone 是不正确的

		var upstreamReq = CloneReqDeep(request, ctx)
		var toHost = request.Host
		fmt.Fprintf(pbuf, "---req %v----------from %v to %v----------------------------------------\n",
			count, u.URL(), toHost)
		WithDumpReq(pbuf)(request)
		var upstreamClient = &http.Client{
			Transport:     trans,
			CheckRedirect: nil,
			Jar:           nil,
			Timeout:       0,
		}
		if upstreamResp, err = upstreamClient.Do(upstreamReq); err != nil {
			fmt.Fprintf(pbuf, "failed request to upstream error: <%T>%v\n", err, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer upstreamResp.Body.Close()
		fmt.Fprintf(pbuf, "---resp %v---------------------------------------------------------------\n",
			count)
		WithDumpResp(pbuf)(upstreamResp)

		// 这个方法不是 pipe 作用，不符合预期
		// _ = resp.Write(w)
		if err = PipeResponse(upstreamResp, w); err != nil {
			fmt.Fprintf(pbuf, "failed write client response from upstream response")
		}
	}
}

type userContextT struct {
	seq            int64
	serverUrl      string
	upstreamDomain string
	upstreamIpPort string
	reqText        []byte
	respText       []byte
}

// String 表示，为了美观打印
func (u *userContextT) String() string {
	var w = bytes.NewBufferString("")
	fmt.Fprintf(w, "---req %v----------from %v to %v %v----------------------------------------\n",
		u.seq, u.serverUrl, u.upstreamDomain, u.upstreamIpPort)
	fmt.Fprintf(w, "%s%s", u.reqText, u.respText)
	return w.String()
}

const pipeHTTPSProxyUserCtxKeyName = "pipeHTTPSProxyUserCtxKeyName"

func getUserCtx(ctx context.Context) *userContextT {
	var vif = ctx.Value(pipeHTTPSProxyUserCtxKeyName)
	var v, ok = vif.(*userContextT)
	if ok && v != nil {
		return v
	}
	return nil
}

func setUserCtx(ctx context.Context, v *userContextT) context.Context {
	return context.WithValue(ctx, pipeHTTPSProxyUserCtxKeyName, v)
}

func getReqText(req *http.Request) []byte {
	var w = bytes.NewBufferString("")
	WithDumpReq(w)(req)
	return w.Bytes()
}

func getRespText(resp *http.Response) []byte {
	var w = bytes.NewBufferString("")
	WithDumpResp(w)(resp)
	return w.Bytes()
}

// GetServerHandleFuncV2 换个思路，直接使用现成的库完成转发
func GetServerHandleFuncV2(ctx context.Context, trans *http.Transport, httpErrLog *log.Logger, seq *atomic.Int64, u Url) http.HandlerFunc {
	// ServeHTTP 将会把请求 pipe 到 upstream

	// 注意错误：Rewrite 和 Director 不能同时都设置，也不能同时都不设置
	var fnRewrite = func(prxReq *httputil.ProxyRequest) {
		// 警告：不要在这里使用 request 克隆，重新提供 context 最好的函数就是 withContext，不是 Clone，会更重
		// 警告：prxReq.In 和 prxReq.Out 共享 body ，如果在 In 中读取完毕重置了，也要在 Out 中重置，否则 Out 将读取不到 body

		// 必须设置，否则没有会报错
		prxReq.Out.URL.Scheme = u.Scheme     // reverseproxy.go:664: http: proxy error: unsupported protocol scheme ""
		prxReq.Out.URL.Host = prxReq.In.Host // reverseproxy.go:664: http: proxy error: http: no Host in request URL

	}
	var fnModifyResponse = func(response *http.Response) error {
		var userCtx = getUserCtx(response.Request.Context())
		//WithDumpReq(w)(response.Request) // 不能在这里读取，因为 body 已经 close
		userCtx.respText = getRespText(response)
		fmt.Fprint(os.Stdout, userCtx.String())
		return nil
	}
	var rev = httputil.ReverseProxy{
		Rewrite:        fnRewrite, // 入参是收到 client 的请求
		Director:       nil,       // 入参已经是发送给 upstream 的请求了
		Transport:      trans,
		FlushInterval:  0,
		ErrorLog:       httpErrLog,
		BufferPool:     nil,
		ModifyResponse: fnModifyResponse,
		ErrorHandler:   nil,
	}
	return func(w http.ResponseWriter, req *http.Request) {
		// 在这里获取 request 比较理想，比 httputil.ReverseProxy.Rewrite 更好，不会影响到 body 的再次读取不到
		var userCtx = &userContextT{
			serverUrl:      u.URL(),
			seq:            seq.Add(1),
			upstreamDomain: req.Host,
			upstreamIpPort: "", // 空的字段也要填写，可以知道所有字段
			reqText:        nil,
			respText:       nil,
		}
		userCtx.reqText = getReqText(req)
		ctx = setUserCtx(ctx, userCtx)
		req = req.WithContext(ctx)
		rev.ServeHTTP(w, req)
	}
}

// 这里留下来调试，按需使用，可以查看直接发送给 upstream 的请求和响应

type internal struct {
	trans *http.Transport
}

func (i internal) RoundTrip(req *http.Request) (*http.Response, error) {
	WithDumpReq(os.Stdout)(req)
	var r, err = i.trans.RoundTrip(req)
	if err == nil {
		WithDumpResp(os.Stdout)(r)
	}
	return r, err
}
