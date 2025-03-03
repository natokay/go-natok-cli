package conf

import (
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

var AppConf *AppConfig

// AppConfig 应用配置
type AppConfig struct {
	Natok Natok `yaml:"natok"`
}

type Natok struct {
	Server        []Server `yaml:"server"`          //服务器端
	CertKeyPath   string   `yaml:"cert-key-path"`   //密钥路径
	CertPemPath   string   `yaml:"cert-pem-path"`   //证书路径
	LogFilePath   string   `yaml:"log-file-path"`   //日志路径
	LogDebugLevel bool     `yaml:"log-debug-level"` //Debug日志
}

// Server NATOK服务配置
type Server struct {
	InetHost  string `yaml:"host"`       // 服务器地址
	InetPort  int    `yaml:"port"`       // 服务器端口
	AccessKey string `yaml:"access-key"` //访问秘钥

}

// AppConfig Load 加载配置
func init() {
	baseDir := getCurrentAbPath()
	// 读取文件内容
	file, err := os.ReadFile(baseDir + "conf.yaml")
	if err != nil {
		log.Error(err)
		panic(err)
	}
	// 利用json转换为AppConfig
	appConfig := new(AppConfig)
	err = yaml.Unmarshal(file, appConfig)
	if err != nil {
		log.Error(err)
		panic(err)
	}
	conf := &appConfig.Natok
	compile, err := regexp.Compile("^/|^\\\\|^[a-zA-Z]:")
	// 密钥文件
	if conf.CertKeyPath != "" && !compile.MatchString(conf.CertKeyPath) {
		log.Infof("%s -> %s", conf.CertKeyPath, baseDir+conf.CertKeyPath)
		conf.CertKeyPath = baseDir + conf.CertKeyPath
	}
	// 证书文件
	if conf.CertPemPath != "" && !compile.MatchString(conf.CertPemPath) {
		log.Infof("%s -> %s", conf.CertPemPath, baseDir+conf.CertPemPath)
		conf.CertPemPath = baseDir + conf.CertPemPath
	}

	// 日志记录配置
	log.SetLevel(log.InfoLevel)
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		ForceColors:     true,
		TimestampFormat: "2006-01-02 15:04:05.000",
	})
	// 在输出日志中添加文件名和方法信息
	if appConfig.Natok.LogDebugLevel {
		log.SetReportCaller(true)
		log.SetLevel(log.DebugLevel)
	}
	// 日志记录输出文件
	if conf.LogFilePath != "" && !compile.MatchString(conf.LogFilePath) {
		log.Infof("%s -> %s", conf.LogFilePath, baseDir+conf.LogFilePath)
		conf.LogFilePath = baseDir + conf.LogFilePath
		logFile, err := os.OpenFile(conf.LogFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatal(err)
		} else {
			// 组合一下即可，os.Stdout代表标准输出流
			multiWriter := io.MultiWriter(logFile, os.Stdout)
			log.SetOutput(multiWriter)
		}
	}
	AppConf = appConfig
}

// 最终方案-全兼容
func getCurrentAbPath() string {
	dir := getCurrentAbPathByExecutable()
	if strings.Contains(dir, getTmpDir()) {
		dir = getCurrentAbPathByCaller()
	}
	return dir + "/"
}

// 获取系统临时目录，兼容go run
func getTmpDir() string {
	dir := os.Getenv("TEMP")
	if dir == "" {
		dir = os.Getenv("TMP")
	}
	res, _ := filepath.EvalSymlinks(dir)
	return res
}

// 获取当前执行文件绝对路径
func getCurrentAbPathByExecutable() string {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	res, _ := filepath.EvalSymlinks(filepath.Dir(exePath))
	return res
}

// 获取当前执行文件绝对路径（go run）
func getCurrentAbPathByCaller() string {
	var abPath string
	if _, filename, _, ok := runtime.Caller(0); ok {
		if lst := strings.LastIndex(filename, "/conf"); lst != -1 {
			abPath = filename[0:lst]
		}
	}
	return abPath
}
