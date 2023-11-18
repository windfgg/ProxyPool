# ProxyPool

ProxyPool是一个使用Go编写的透明MITM代理转发库，旨在将请求的流量通过GoProxy透明地转发到代理池上。

## 主要特性：

- 透明的MITM代理转发：该库能够截获传入的请求流量，并将其转发到预定义的代理池上，实现透明的中间人攻击（Man-in-the-Middle）代理转发。
- 支持HTTPS流量：该库能够处理HTTPS流量，并在转发过程中保持流量的加密性。
- 灵活的配置选项：通过库提供的配置选项，你可以自定义获取代理地址、拉取代理间隔时间、代理过期时间、证书设置等，以满足不同的需求。
- 高性能：该库经过优化，具有较高的性能和吞吐量，能够处理大量的并发请求。

## 安装

运行程序会在当前所在目录创建`conf`文件夹，首次运行会在文件夹内生成CA证书和配置文件，程序运行中将使用该证书。

### 配置模板

```yaml
ProxyUrl: ""
ExpTime: 25
IntervalTime: 3
Auth:
  UserName: ""
  Password: ""
DetailLog: false
```

### Docker

```shell
docker run -p 9876:8080 \
    -v $(pwd)/proxy-pool/conf:/conf \
    ghcr.io/windfgg/proxypool/proxy-pool:latest
```

### 二进制发布版

[Releases页面](https://github.com/windfgg/ProxyPool/releases)
