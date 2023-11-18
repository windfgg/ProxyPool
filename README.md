# 信任证书
```

//openssl pkcs12 -in NolanProxyPoolrootCert.pfx -out narkpool.pem -nodes

sudo cp cert.pem /usr/local/share/ca-certificates/cert.crt

sudo update-ca-certificates

openssl verify -CAfile /etc/ssl/certs/ca-certificates.crt cert.pem

```