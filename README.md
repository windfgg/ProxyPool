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
ProxyUrl: ""            //代理URL
ExpTime: 25             //代理过期时间
IntervalTime: 3         //拉取间隔
Auth:                   //是否鉴权
  UserName: ""          //鉴权用户名
  Password: ""          //鉴权密码
DetailLog: false        //输出详细日志
DetailLogRequest: false //输出请求的详细日志
MaxConnect: 1000        //最大并发数量 用来限制CPU占用
IsCertStore: true       //是否存储证书 用来控制内存占用
```

### Docker

```shell
1. 创建文件夹
mkdir ProxyPool && cd ProxyPool
2. 运行Docker 镜像
docker run -d -p 9876:8080 --name ProxyPool \
    --restart always \
    -v $(pwd):/conf \
    ghcr.io/windfgg/proxypool/proxy-pool:latest
3.修改目录下的 config.yml 具体参数请参考配置模板
4.安装证书 (windos启动的代理池需要手动信任 ProxyPool.crt 证书)
sudo cp $(pwd)/ProxyPool.crt /usr/local/share/ca-certificates/
sudo update-ca-certificates
//查看证书是否信任
openssl verify -CAfile /etc/ssl/certs/ca-certificates.crt ProxyPool.crt
```

### 二进制发布版

[Releases页面](https://github.com/windfgg/ProxyPool/releases)
