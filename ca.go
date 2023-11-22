package main

import (
	cryptoRand "crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"os"
	"path/filepath"
	"time"

	"github.com/elazarl/goproxy"
)

// CreteCA
func LoadProxyCA() {
	pwd, _ := os.Getwd()
	CaCertPath := filepath.Join(pwd, "conf", "ProxyPool.pem")
	CaKeyPath := filepath.Join(pwd, "conf", "ProxyPoolKey.pem")

	_, caExist := os.Stat(CaCertPath)

	if os.IsNotExist(caExist) {
		caCert, caKey, _ := generateCACertificate()
		saveCertificateToFile(caCert, caKey, CaCertPath, CaKeyPath)
	}
}

// 生成CA证书
//
//	@return []byte
//	@return []byte
//	@return error
func generateCACertificate() ([]byte, []byte, error) {
	// 生成私钥
	caKey, err := rsa.GenerateKey(cryptoRand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	// 构建证书模板
	caTemplate := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "ProxyPool"},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0), // 有效期为 10 年
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	// 使用模板生成证书
	caCert, err := x509.CreateCertificate(cryptoRand.Reader, &caTemplate, &caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		return nil, nil, err
	}

	return caCert, x509.MarshalPKCS1PrivateKey(caKey), nil
}

// 保存CA证书到文件
//
//	@param cert
//	@param key
//	@param filePath
//	@return error
func saveCertificateToFile(cert []byte, key []byte, certPath string, keyPath string) {
	os.WriteFile(certPath, cert, 0644)
	os.WriteFile(keyPath, key, 0644)
}

// go proxy 设置 ca证书
//
//	@param caCert
//	@param caKey
//	@return error
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
