package main

import (
	"log"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/jinzhu/configor"
	"gopkg.in/yaml.v2"
)

func LoadConfig() {
	pwd, _ := os.Getwd()
	ConfigPath := filepath.Join(pwd, "conf", "config.yml")
	os.MkdirAll("./conf", 0755)

	// 检查配置文件是否存在
	_, err1 := os.Stat(ConfigPath)
	if os.IsNotExist(err1) {
		data, _ := yaml.Marshal(Config{
			ProxyUrl:     "",
			ExpTime:      25,
			IntervalTime: 3,
			Auth: Auth{
				UserName: "",
				Password: "",
			},
			DetailLog: false,
		})
		os.Create(ConfigPath)
		os.WriteFile(ConfigPath, data, 0644)
		log.Println("Frist Run, Init config.yml...")
	} else if err1 == nil {
		log.Println("Load config.yml")
	} else {
		log.Println("Load Config Error:", err1)
	}

	configor.Load(&config, "conf/config.yml")

	url, err2 := url.Parse(config.ProxyUrl)
	if config.ProxyUrl == "" {
		log.Println("Please Fill ProxyUrl to Config:")
		time.Sleep(3 * time.Second)
		os.Exit(-1)
	} else if err2 != nil {
		log.Println("Failed to parse proxy URL:", url)
		time.Sleep(3 * time.Second)
		os.Exit(-1)
	}
}
