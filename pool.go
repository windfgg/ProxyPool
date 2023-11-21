package main

import (
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
)

var proxies sync.Map

// 加载代理工厂
func LoadProxyFactory() {
	// 加载代理的Ticker
	loadTicker := time.NewTicker(time.Duration(config.IntervalTime) * time.Second)
	//defer loadTicker.Stop()

	// 删除过期代理的Ticker
	removeTicker := time.NewTicker(time.Duration(config.IntervalTime) * time.Second)
	//defer removeTicker.Stop()

	// 启动一个 goroutine 来处理定时器事件
	go func() {
		for {
			select {
			case <-loadTicker.C:
				PullProxy()
			case <-removeTicker.C:
				RemoveExpProxy()
			}
		}
	}()
}

// 获取随机代理
//
//	@return string
func GetRandomPorxy() string {
	// 设置随机数种子
	rand.NewSource(time.Now().UnixNano())

	// 随机获取一个键
	var randomKey string
	proxies.Range(func(key, value interface{}) bool {
		if rand.Intn(2) == 0 {
			randomKey = key.(string)
			return false // 停止遍历
		}
		return true
	})
	//proxies.Delete(randomKey)
	return randomKey
}

// 拉取代理
func PullProxy() {
	client := resty.New()
	resp, err := client.R().
		Get(config.ProxyUrl)
	if err != nil {
		log.Print("Pull Proxy Error:", err)
	}
	var storeCout int

	currentTime := time.Now()
	newTime := currentTime.Add(time.Duration(config.ExpTime) * time.Second)

	proxyArray := strings.Split(string(resp.Body()), "\n")
	for _, proxy := range proxyArray {
		_, found := proxies.Load(strings.Replace(proxy, "\r", "", -1))
		if !found {
			if !IsInternalIP(strings.Split(proxy, ":")[0]) {
				proxies.Store(strings.Replace(proxy, "\r", "", -1), newTime)
				storeCout++
			}
		}
	}

	if storeCout != 0 {
		log.Printf("Loaded %v new proxies. Total: %v.", storeCout, GetProxiesCount())
	}
}

// GetProxiesCount
// 获取代理池数量
//
//	@return int 数量
func GetProxiesCount() int {
	count := 0
	proxies.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

// 删除过期代理
func RemoveExpProxy() {
	now := time.Now()

	var expiredProxies []string
	proxies.Range(func(key, value interface{}) bool {
		v, _ := value.(time.Time)
		k, _ := key.(string)
		if v.Before(now) {
			expiredProxies = append(expiredProxies, k)
		}

		return true
	})

	for _, proxy := range expiredProxies {
		proxies.Delete(proxy)
	}
	if len(expiredProxies) != 0 && config.DetailLog {
		log.Printf("Remove %v expired proxies. Total: %v.", len(expiredProxies), GetProxiesCount())
	}
}
