# 信任证书
```
// 使用opensll 生成证书
openssl pkcs12 -in cert.pfx -out cert.pem -nodes

//将证书一到系统里
sudo cp cert.pem /usr/local/share/ca-certificates/cert.crt

//更新证书
sudo update-ca-certificates

//验证证书
openssl verify -CAfile /etc/ssl/certs/ca-certificates.crt cert.pem

```
# Docker
```
docker run -p 8080:8080 \
    -v /root/PorxyPool/config.yml:/config.yml \
    ghcr.io/windfgg/proxypool/proxy-pool:latest
```

# config.yml
```
ProxyUrl: ""
ExpTime: 25
IntervalTime: 3
Auth:
  UserName: ""
  Password: ""
DetailLog: false

```
