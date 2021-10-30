package conf

import (
	"encoding/json"
	"io/ioutil"
)

var AppConf *AppConfig

// AppConfig 应用配置
type AppConfig struct {
	Server Server `json:"natok.server"`
}

// Server NATOK服务配置
type Server struct {
	InetHost  string `json:"host"`       // 服务器地址
	InetPort  int    `json:"port"`       // 服务器端口
	AccessKey string `json:"access-key"` //访问秘钥

	CertKeyPath string `json:"cert-key-path"`
	CertPemPath string `json:"cert-pem-path"`
}

// AppConfig Load 加载配置
func init() {
	// 读取文件内容
	file, err := ioutil.ReadFile("application.json")
	if err != nil {
		panic(err)
	}
	// 利用json转换为AppConfig
	appConfig := new(AppConfig)
	err = json.Unmarshal(file, appConfig)
	if err != nil {
		panic(err)
	}
	AppConf = appConfig
}
