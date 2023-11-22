package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/elazarl/goproxy"
)

func main() {
	log.Println("Application Start")

	defer func() {
		if r := recover(); r != nil {
			log.Println("Gload Error:", r)
		}
		log.Printf("Application Exit \n")
		time.Sleep(3 * time.Second)
	}()

	LoadConfig()
	LoadProxyCA()
	LoadProxyFactory()

	time.Sleep(3 * time.Second)
	verbose := flag.Bool("v", config.DetailLog, "记录发送到代理的每个请求的信息")
	addr := flag.String("addr", ":8080", "代理监听地址和端口")

	proxy := goproxy.NewProxyHttpServer()
	proxy.NonproxyHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte(strconv.Itoa(GetProxiesCount())))
	})
	proxy.Logger = &customLogger{log.New(os.Stderr, "", log.LstdFlags)}
	proxy.Verbose = *verbose
	SetOnRequest(proxy)

	log.Printf("Starting Proxy Pool in [ %v ]\n", *addr)
	http.ListenAndServe(*addr, proxy)
}

func SetOnRequest(proxy *goproxy.ProxyHttpServer) {
	proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)
	proxy.OnResponse().DoFunc(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		if resp.StatusCode == 403 || resp.StatusCode == 493 {
			if config.DetailLog {
				log.Printf("Trigger 403 or 439. Remove Proxies: [%v].", ctx.UserData)
			}
			proxies.Delete(ctx.UserData)
		}
		return resp
	})

	proxy.OnRequest().DoFunc(func(request *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		proxy := GetRandomPorxy()
		if proxy == "" {
			if config.DetailLog {
				log.Printf("Incorrect pull to proxy,Addres %v", request.URL)
			}
			return request, nil
		}

		if IsInternalIP(request.Host) {
			if config.DetailLog {
				log.Printf("Proxy: %v ---> Addres: %v", "Intranet IP", request.URL)
			}
			return request, nil
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
		ctx.UserData = proxy

		if config.DetailLog {
			log.Printf("Proxy: %v ---> Addres: %v", proxy, request.URL.Host)
		}

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
