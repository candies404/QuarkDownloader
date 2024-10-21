package util

import (
	"QuarkDownloader/config"
	"bytes"
	"encoding/json"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

var (
	delay int
	proxy string
)

func init() {
	delay = int(config.Cfg.Delay * 1000)
	proxy = config.Cfg.Proxy
}

// GetHTTPClient 根据代理配置生成 http.Client
func GetHTTPClient() (*http.Client, error) {
	// 创建默认的 Transport
	transport := &http.Transport{}

	// 如果 proxy 不为空，则设置代理
	if proxy != "" {
		proxyURL, err := url.Parse(proxy)
		if err != nil {
			return nil, err
		}
		// 设置代理
		transport.Proxy = http.ProxyURL(proxyURL)
	}

	// 返回配置好的 http.Client
	client := &http.Client{
		Timeout:   60 * time.Second, // 设置超时
		Transport: transport,
	}

	return client, nil
}

// SendRequest 发送HTTP请求的通用方法
func SendRequest(method, reqURL string, params map[string]string, data interface{}, headers map[string]string) (*http.Response, error) {
	// 随机延迟，避免过快的请求发送
	time.Sleep(time.Duration(rand.Intn(delay)) * time.Millisecond)

	// 创建一个HTTP客户端，设置超时
	client, err := GetHTTPClient()
	if err != nil {
		return nil, err
	}

	// 如果有URL参数，进行拼接并进行URL编码（仅对值进行编码）
	if len(params) > 0 {
		queryParams := url.Values{}
		for key, value := range params {
			// 对 value 进行 URL 编码，而不是 key
			queryParams.Add(key, value)
		}
		// 拼接编码后的参数
		reqURL = reqURL + "?" + queryParams.Encode()
		//println(reqURL)
	}

	var req *http.Request

	// 根据请求方法（POST/GET）构建请求
	if method == http.MethodPost {
		// 将请求体序列化为JSON
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		// 构建POST请求
		req, err = http.NewRequest(http.MethodPost, reqURL, bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
	} else {
		// 构建GET请求
		req, err = http.NewRequest(http.MethodGet, reqURL, nil)
	}

	if err != nil {
		return nil, err
	}

	// 设置请求头
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// 发送请求并返回响应
	return client.Do(req)
}
