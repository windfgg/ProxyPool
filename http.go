package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

func test() {
	// 设置自定义代理
	proxyURL, _ := url.Parse("http://127.0.0.1:8080")

	// 创建一个带超时的 HTTP 客户端
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			Proxy:           http.ProxyURL(proxyURL),
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	// 并发请求数量
	concurrency := 200

	// 使用 WaitGroup 等待所有请求完成
	var wg sync.WaitGroup
	wg.Add(concurrency)

	startTime := time.Now()

	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()

			// 发起 GET 请求
			resp, err := client.Get("https://api.m.jd.com/")
			if err != nil {
				fmt.Println("Error:", err)
				return
			}
			defer resp.Body.Close()

			// 读取响应内容
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("Error reading response:", err)
				return
			}

			// 打印响应内容长度
			fmt.Printf("Response: %v \n", string(body))
		}()
	}

	// 等待所有请求完成
	wg.Wait()

	// 计算总耗时
	elapsed := time.Since(startTime)
	fmt.Printf("Total time: %s\n", elapsed)
}
