package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/elazarl/goproxy"
)

// LoadProxyCA 加载代理服务器的 CA 证书
func LoadProxyCA() {
	pwd, _ := os.Getwd()
	caCertPath := filepath.Join(pwd, "conf", "ProxyPool.crt")
	caKeyPath := filepath.Join(pwd, "conf", "ProxyPoolKey.pem")

	_, caCertExist := os.Stat(caCertPath)
	_, caKeyExist := os.Stat(caKeyPath)

	if os.IsNotExist(caCertExist) || os.IsNotExist(caKeyExist) {
		caCert, caKey, _ := generateCACertificate()
		saveCertificateToFile(caCert, caKey, caCertPath, caKeyPath)
		if runtime.GOOS == "linux" {
			installCertificateLinuxErr := installCertificateLinux(caCertPath)
			if installCertificateLinuxErr != nil {
				log.Println("InstallCertificate Error:", installCertificateLinuxErr)
			} else {
				log.Println("InstallCertificate Success")
			}
		}
	}
}

// 安装证书到 Linux 系统
func installCertificateLinux(certPath string) error {
	cmd := exec.Command("sudo", "cp", certPath, "/usr/local/share/ca-certificates/ProxyPool.crt")
	err := cmd.Run()
	if err != nil {
		return err
	}

	cmd = exec.Command("sudo", "update-ca-certificates")
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

// 生成 CA 证书
func generateCACertificate() ([]byte, []byte, error) {
	// 生成私钥
	caKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, err
	}

	// 构建证书模板
	caTemplate := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "WindfggProxyPool"},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0), // 有效期为 10 年
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	// 使用模板生成证书
	caCert, err := x509.CreateCertificate(rand.Reader, &caTemplate, &caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		return nil, nil, err
	}

	return caCert, x509.MarshalPKCS1PrivateKey(caKey), nil
}

// 保存 CA 证书到文件
func saveCertificateToFile(cert []byte, key []byte, certPath string, keyPath string) {
	certFile, _ := os.Create(certPath)
	keyFile, _ := os.Create(keyPath)

	pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: cert})
	pem.Encode(keyFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: key})

	defer keyFile.Close()
	defer certFile.Close()
}

// 设置 CA 证书
func SetCA(caCert, caKey []byte) {
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	goproxyCa, _ := tls.X509KeyPair(caCert, caKey)

	goproxy.GoproxyCa = goproxyCa
	goproxy.OkConnect = &goproxy.ConnectAction{Action: goproxy.ConnectAccept, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	goproxy.MitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectMitm, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	goproxy.HTTPMitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectHTTPMitm, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	goproxy.RejectConnect = &goproxy.ConnectAction{Action: goproxy.ConnectReject, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
}
