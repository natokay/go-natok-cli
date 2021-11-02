# go-natok-cli


natok-cli的相关配置：application.json
```shell
{
  "natok.server": {
    "host": "127.0.0.1",            // 服务器地址：域名 或者 ip
    "port": 1001,                   // 服务器端口：可自定义
    "access-key": "",               // 客户端访问密钥，从natok-server的web页面中C端列表里获取
    "cert-key-path": "s-cert.key",  // TSL加密密钥，可自己指定。注：需与server端保持一致
    "cert-pem-path": "s-cert.pem"   // TSL加密证书，可自己指定。注：需与server端保持一致
  }
}
```

**Go 1.13 及以上（推荐）**
```shell
# 配置 GOPROXY 环境变量
go env -w GO111MODULE=on
go env -w GOPROXY=https://goproxy.io,direct
```

构建natok-cli可执行程序

```shell
# 克隆项目
git clone https://github.com/play-sy/go-natok-cli.git

# 进入项目目录
cd go-natok-cli

# 更新/下载依赖
go mod vendor

# 设置目标可执行程序操作系统构架，包括 386，amd64，arm
set GOARCH=amd64

# 设置可执行程序运行操作系统，支持 darwin，freebsd，linux，windows
set GOOS=windows

# cd到main.go目录，打包命令
go build

# 启动程序
./go-natok-cli.exe
```
