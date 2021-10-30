package main

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/kataras/golog"
	"io/ioutil"
	"natok-cli/conf"
	"natok-cli/core"
	"net"
	"strconv"
	"time"
)

// 程序入口
func main() {

	golog.SetLevel("debug")
	golog.Info("Startup natok-cli")

	Start()
}

// Start 启动主服务
func Start() {
	appConf := conf.AppConf.Server
	addr := appConf.InetHost + ":" + strconv.Itoa(appConf.InetPort)
	tlsConf := tlsConfig()
	pooler := &core.PoolHandler{
		Pool: &core.ConnPooler{
			Addr: addr,
			Conf: tlsConf,
		},
		Conns: make([]*core.ConnectHandler, 0, 10),
	}

	connHandler := &core.ConnectHandler{}

	for {
		conn := Connect(addr, tlsConf)
		connHandler.Conn = conn

		msgHandler := &core.NatokServerHandler{
			AccessKey:   appConf.AccessKey,
			PoolHandler: pooler,
			ConnHandler: connHandler,
		}
		msgHandler.HeartBeat()

		connHandler.Listen(conn, msgHandler)
	}
}

// Connect 向NATOK服务段发起连接
func Connect(addr string, conf *tls.Config) net.Conn {
	for {
		var conn net.Conn
		var err error

		if conf != nil {
			conn, err = tls.Dial("tcp", addr, conf)
		} else {
			conn, err = net.Dial("tcp", addr)
		}

		if err != nil {
			golog.Errorf("Conn to natok server error! %+v", err)
			time.Sleep(time.Second * 20)
			continue
		}

		return conn
	}
}

// tlsConfig TSL协议配置
func tlsConfig() *tls.Config {
	tlsConf := conf.AppConf.Server
	cert, err := tls.LoadX509KeyPair(tlsConf.CertPemPath, tlsConf.CertKeyPath)
	if err != nil {
		panic(err)
	}
	certBytes, err := ioutil.ReadFile(tlsConf.CertPemPath)
	if err != nil {
		panic("Unable to read cert.pem")
	}
	clientCertPool := x509.NewCertPool()
	ok := clientCertPool.AppendCertsFromPEM(certBytes)
	if !ok {
		panic("failed to parse root certificate")
	}
	return &tls.Config{
		RootCAs:            clientCertPool,
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
	}
}
