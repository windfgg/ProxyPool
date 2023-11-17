package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/elazarl/goproxy"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/jinzhu/configor"
)

var proxies sync.Map
var config *Config

type Config struct {
	ProxyUrl     string `required:"true" yaml:"ProxyUrl"`
	ExpTime      int    `required:"true" yaml:"ExpTime"`
	IntervalTime int64  `required:"true" yaml:"IntervalTime"`
	Auth         struct {
		UserName string `yaml:"UserName"`
		Password string `yaml:"Password"`
	}
	/*
		Check struct {
			IP            bool   `default:"IP"`
			Url           string `default:"Url"`
			ParallelCount int    `default:"ParallelCount"`
			TimeoutTime   int    `default:"TimeoutTime"`
			IPBeforAdd    bool   `default:"IPBeforAdd"`
		}
	*/

	DetailLog bool `required:"true" yaml:"DetailLog"`
}

func main() {
	configor.Load(&config, "config.yml")
	Init()
}

// 自定义的日志输出结构
type customLogger struct {
	logger *log.Logger
}

// 实现 goproxy.Logger 接口的 Printf 方法
func (c *customLogger) Printf(format string, v ...interface{}) {
	logEntry := fmt.Sprintf(format, v...)

	if !strings.Contains(logEntry, "WARN:") {
		c.logger.Print(logEntry) //过滤掉goproxy 引起的 WARN
	}
}

func Init() {
	var CaCertPath string
	var CaKeyPath string
	ProxyFactory()
	time.Sleep(3 * time.Second)

	if runtime.GOOS == "windows" {
		CaCertPath = "\\ca\\cert.pem"
		CaKeyPath = "\\ca\\key.pem"
	} else {
		CaCertPath = "/ca/cert.pem"
		CaKeyPath = "/ca/key.pem"
	}
	verbose := flag.Bool("v", config.DetailLog, "should every proxy request be logged to stdout") // 设置是否输出连接信息
	addr := flag.String("addr", ":8080", "proxy listen address")                                  // 监听端口和地址

	proxy := goproxy.NewProxyHttpServer()
	logger := log.New(os.Stderr, "", log.LstdFlags)
	proxy.Logger = &customLogger{logger}

	pwd, _ := os.Getwd()
	CaCertPath = pwd + CaCertPath
	CaKeyPath = pwd + CaKeyPath

	caCert, err := os.ReadFile(CaCertPath) // 设置为你刚才生成的证书路径
	if err != nil {
		log.Fatal(err)
	}
	caKey, _ := os.ReadFile(CaKeyPath) // 设置为你刚才生成的证书路径
	SetCA(caCert, caKey)
	proxy.Verbose = *verbose

	/*
		sl, err := net.Listen(*addr,"tcp")
		if err != nil {
			log.Fatal("listen:", err)
		}
	*/
	OnRequest(proxy)
	log.Printf("Starting Proxy Pool %s \n", *addr)
	http.ListenAndServe(*addr, proxy)
}

func SetCA(caCert, caKey []byte) error {
	goproxyCa, err := tls.X509KeyPair(caCert, caKey)
	if err != nil {
		return err
	}
	if goproxyCa.Leaf, err = x509.ParseCertificate(goproxyCa.Certificate[0]); err != nil {
		return err
	}
	goproxy.GoproxyCa = goproxyCa
	goproxy.OkConnect = &goproxy.ConnectAction{Action: goproxy.ConnectAccept, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	goproxy.MitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectMitm, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	goproxy.HTTPMitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectHTTPMitm, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	goproxy.RejectConnect = &goproxy.ConnectAction{Action: goproxy.ConnectReject, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	return nil
}

func IsInternalIP(ipAddress string) bool {
	if ipAddress == "" {
		return false
	}

	internalIPs := []string{"127.0.0.1", "10.", "172.16.", "172.17.", "172.18.", "172.19.", "172.20.", "172.21.", "172.22.", "172.23.", "172.24.", "172.25.", "172.26.", "172.27.", "172.28.", "172.29.", "172.30.", "172.31.", "192.168."}

	for _, internalIP := range internalIPs {
		if strings.HasPrefix(ipAddress, internalIP) {
			return true
		}
	}

	return false
}

func OnRequest(proxy *goproxy.ProxyHttpServer) {
	proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)
	proxy.OnResponse().DoFunc(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		if resp.StatusCode == 403 || resp.StatusCode == 493 {
			log.Printf("Trigger 403 or 439. Remove Proxies: [%v].", resp.Proto)
		}
		return resp
	})

	proxy.OnRequest().DoFunc(func(request *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		proxy := GetRandomPorxy()
		if proxy == "" {
			log.Printf("Incorrect pull to proxy,Addres %v", request.URL)
			return request, ctx.Req.Response
		}

		if IsInternalIP(request.Host) {
			log.Printf("Proxy: %v Addres: %v", "No Proxy", request.URL)
			return request, ctx.Req.Response
		}

		// 设置代理地址
		proxyURL, err := url.Parse("http://" + proxy)
		if err != nil {
			log.Println("Failed to parse proxy URL:", err)
			return request, nil
		}
		// 创建自定义的 Transport，并设置代理
		transport := &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // 如果代理使用的是自签名证书，可能需要跳过证书验证
			},
		}
		//if config.DetailLog {
		log.Printf("Proxy: %v ---> Addres: %v", proxy, request.URL.Host)
		//}

		response, err := transport.RoundTrip(request)
		if err != nil {
			if config.DetailLog {
				log.Printf("[%v] Failed to send request via proxy: %v", proxy, err)
			}

			return request, nil
		}

		return request, response
	})
}

// 获取随机代理
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

// 代理工厂
func ProxyFactory() {
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

// 拉取代理
func PullProxy() {
	client := resty.New()
	resp, err := client.R().
		Get(config.ProxyUrl)
	if err != nil {
		log.Fatalln("Pull Proxy Error:", err)
		return
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

func CreateHttpServer() {
	router := gin.Default()

	router.GET("/GetProxies", GetProxies)

	err := router.Run(":5000")
	if err != nil {
		fmt.Println("Failed to start server:", err)
	}
}

type Proxy struct {
	Key   string    `json:"key"`
	Value time.Time `json:"value"`
}

func GetProxies(c *gin.Context) {
	var proxyList []Proxy

	proxies.Range(func(key, value interface{}) bool {
		k, _ := key.(string)
		v, _ := value.(time.Time)
		proxy := Proxy{
			Key:   k,
			Value: v,
		}
		proxyList = append(proxyList, proxy)

		return true
	})

	c.JSON(200, gin.H{
		"sta":  true,
		"list": proxyList,
	})
}
