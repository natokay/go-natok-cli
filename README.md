# NATOK

- natok是一个将局域网内个人服务代理到公网可访问的内网穿透工具，基于tcp协议、支持udp协议，支持任何tcp上层协议（列如：http、https、ssh、telnet、data base、remote desktop....）。
- 目前市面上提供类似服务的有：花生壳、natapp、ngrok等等。当然，这些工具都很优秀！但是免费提供的服务都很有限，想要有比较好的体验都需要支付一定的套餐费用，由于数据包会流经第三方，因此总归有些不太友好。
- natok-server与natok-cli都基于GO语言开发，几乎不存在并发问题。运行时的内存开销也很低，一般在几十M左右。所以很推荐自主搭建服务！


**服务端与客户端**

| 服务                     |支持系统| 下载地址                                               |
| ------------------------|----- | ------------------------------------------------------ |
| natok-cli |linux/windows| [GitHub](https://github.com/natokay/go-natok-cli/releases) |
| natok-server| linux/windows|[GitHub](https://github.com/natokay/go-natok-server/releases) |



**natok-cli的相关配置：conf.yaml**
```yaml
natok:
  server:
    - host: natok1.cn #服务器地址：域名 或者 ip
      port: 1001      #服务器端口：可自定义
      #客户端访问密钥，从natok-server的web页面中C端列表里获取
      access-key: 74a7a42fcdc4ccb6c8641ce543fe2e07
    - host: natok2.cn
      port: 1001
      access-key: 74a7a42fcdc4ccb6c8641ce543fe2e07
  cert-key-path: s-cert.key #TSL加密密钥，可自己指定。注：需与server端保持一致
  cert-pem-path: s-cert.pem #TSL加密证书，可自己指定。注：需与server端保持一致
  log-file-path: out.log    #程序日志输出配置
```

- windows系统启动： 双击 natok-cli.exe
```powershell
# 注册服务，自动提取管理员权限：
natok-cli.exe install
# 卸载服务，自动提取管理员权限：
natok-cli.exe uninstall
# 启停服务，自动提取管理员权限：
natok-cli.exe start/stop
# 启停服务，终端管理员权限
net start/stop natok-cli
```
- Linux系统启动：
```shell
# 授予natok-cli可执权限
chmod 755 natok-cli
# 启动应用
nohup ./natok-cli > /dev/null 2>&1 &
```

---
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
go mod tidy
go mod vendor

# 设置目标可执行程序操作系统构架，包括 386，amd64，arm
go env -w GOARCH=amd64

# 设置可执行程序运行操作系统，支持 darwin，freebsd，linux，windows
go env -w GOOS=windows

# golang windows 程序获取管理员权限(UAC)
rsrc -manifest nac.manifest -o nac.syso

# cd到main.go目录，打包命令
go build

# 启动程序
./natok-cli.exe
```

## 版本描述
**natok:1.0.0**
natok-cli与natok-server网络代理通信基本功能实现。

**natok:1.1.0**
natok-cli与natok-server支持windows平台注册为服务运行，可支持开机自启，保证服务畅通。

**natok:1.2.0**
natok-cli可与多个natok-server保持连接，支持从多个不同的natok-server来访问natok-cli，以实现更快及更优的网络通信。

**natok:1.3.0**
natok-cli与natok-server可支持udp网络代理。
