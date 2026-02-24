package ihttp

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"
)

// 全局复用 Transport（连接池核心），避免每次创建新连接
var sharedTransport = &http.Transport{
	MaxIdleConns:        32, // 最大空闲连接数（默认2）
	MaxIdleConnsPerHost: 8,
	IdleConnTimeout:     90 * time.Second, // 空闲连接超时时间（默认无）
	MaxConnsPerHost:     0,               // 每个主机最大并发连接数
	TLSHandshakeTimeout: 5 * time.Second,  // TLS握手超时
	DisableCompression:  false,            // 启用gzip压缩（节省带宽）
	DialContext: (&net.Dialer{
		Timeout:   5 * time.Second,  // 拨号超时
		KeepAlive: 30 * time.Second, // TCP保活时间
	}).DialContext,
}
var sharedClient = &http.Client{
	Transport: sharedTransport,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 { // 限制重定向次数（默认10次，避免无限重定向）
			return errors.New("too many redirects (max 5)")
		}
		return nil
	},
}

func Get(link string, timeout time.Duration, headers map[string]string) (int, []byte, error) {
	return _do(http.MethodGet, link, nil, timeout, headers)
}
func Post(link string, pl []byte, timeout time.Duration, headers map[string]string) (int, []byte, error) {
	return _do(http.MethodPost, link, pl, timeout, headers)
}
func Delete(link string, timeout time.Duration, headers map[string]string) (int, []byte, error) {
	return _do(http.MethodDelete, link, nil, timeout, headers)
}
func Put(link string, timeout time.Duration, headers map[string]string) (int, []byte, error) {
	return _do(http.MethodPut, link, nil, timeout, headers)
}
func _do(method, link string, pl []byte, timeout time.Duration, headers map[string]string) (int, []byte, error) {
	buffer := bytes.NewBuffer(pl)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, method, link, buffer)
	if err != nil {
		return 0, nil, errors.New("create request failed: " + link + ", err: " + err.Error())
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := sharedClient.Do(req)
	defer func() {
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close() // 忽略关闭错误（核心是确保关闭）
		}
	}()
	if err != nil {
		var timeoutErr error
		if uErr, ok := err.(*url.Error); ok {
			if netErr, ok := uErr.Err.(net.Error); ok {
				if netErr.Timeout() {
					timeoutErr = errors.New("request timeout: " + link + ", timeout: " + timeout.String())
				} else if netErr.Temporary() {
					timeoutErr = errors.New("temporary error (network issue): " + link + ", err: " + netErr.Error())
				}
			}
		}
		if timeoutErr != nil {
			return 0, nil, timeoutErr
		}
		return 0, nil, errors.New("request failed: " + link + ", err: " + err.Error())
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, errors.New("read response body failed: " + link + ", err: " + err.Error())
	}

	return resp.StatusCode, body, nil
}
