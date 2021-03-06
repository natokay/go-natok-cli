# NATOK

- natok是一个将局域网内个人服务代理到公网可访问的内网穿透工具，基于tcp协议，支持任何tcp上层协议（列如：http、https、ssh、telnet、data base、remote desktop....）。
- 目前市面上提供类似服务的有：花生壳、natapp、ngrok等等。当然，这些工具都很优秀！但是免费提供的服务都很有限，想要有比较好的体验都需要支付一定的套餐费用，由于数据包会流经第三方，因此总归有些不太友好。
- natok-server与natok-cli都基于GO语言开发，几乎不存在并发问题。运行时的内存开销也很低，一般在几十M左右。所以很推荐自主搭建服务！


**服务端与客户端**

| 服务                     |支持系统| 下载地址                                               |
| ------------------------|----- | ------------------------------------------------------ |
| natok-cli |linux/windows| [GitHub](https://github.com/natokay/go-natok-cli/releases) |
| natok-server| linux/windows|[GitHub](https://github.com/natokay/go-natok-server/releases) |


# go-natok-cli

natok-cli的相关配置：application.json
```json5
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
git clone https://github.com/natokay/go-natok-cli.git

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
