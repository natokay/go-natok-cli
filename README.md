# go-natok-cli


自主构建natok-cli可执行程序

```
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
