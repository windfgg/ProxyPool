package main

import (
	"fmt"
	"log"
	"strings"
)

var config *Config

type Config struct {
	ProxyUrl     string `required:"true" yaml:"ProxyUrl"`
	ExpTime      int    `required:"true" yaml:"ExpTime"`
	IntervalTime int64  `required:"true" yaml:"IntervalTime"`
	Auth         Auth   `yaml:"Auth"`

	DetailLog        bool `required:"true" yaml:"DetailLog"`
	DetailLogRequest bool `required:"true" yaml:"DetailLogRequest"`

	MaxConnect  int  `required:"true" yaml:"MaxConnect"`
	IsCertStore bool `required:"true" yaml:"IsCertStore"`
}

type Auth struct {
	UserName string `yaml:"UserName"`
	Password string `yaml:"Password"`
}

// 自定义的日志输出结构
type customLogger struct {
	logger *log.Logger
}

func (c *customLogger) Printf(format string, v ...interface{}) {
	logEntry := fmt.Sprintf(format, v...)

	if !strings.Contains(logEntry, "WARN:") {
		c.logger.Print(logEntry) //过滤掉goproxy 引起的 WARN
	}
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
