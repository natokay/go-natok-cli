package main

import (
	"crypto/tls"
	"crypto/x509"
	"natok-cli/conf"
	"natok-cli/core"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/kardianos/service"
	"github.com/kataras/golog"
)

type Program struct{}

func (p *Program) Start(s service.Service) error {
	go p.run()
	return nil
}

func (p *Program) run() {
	golog.Info("Started natok client service")
	Start()
}

func (p *Program) Stop(s service.Service) error {
	golog.Info("Stop natok client service")
	return nil
}

func init() {
	// 日志记录处理
	golog.SetLevel("debug")
	if conf.AppConf.Natok.LogFilePath != "" {
		logFile, err := os.OpenFile(conf.AppConf.Natok.LogFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			golog.Error(err)
		} else {
			golog.AddOutput(logFile)
		}
	}
}

// 程序入口
func main() {
	svcConfig := &service.Config{
		Name:        "natok-cli",
		DisplayName: "Natok Client Service",
		Description: "Go语言实现的内网代理客户端服务",
	}

	prg := &Program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		golog.Fatal(err)
	}

	if len(os.Args) > 1 {
		if os.Args[1] == "install" {
			if se := s.Install(); se != nil {
				golog.Errorf("Service installation failed. %+v", se)
			} else {
				golog.Info("Service installed")
			}
			return
		}
		if os.Args[1] == "uninstall" {
			if se := s.Uninstall(); se != nil {
				golog.Errorf("Service uninstall failed. %+v", se)
			} else {
				golog.Info("Service uninstalled")
			}
			return
		}
		if os.Args[1] == "start" {
			if se := s.Start(); se != nil {
				golog.Errorf("Service start failed. %+v", se)
			} else {
				golog.Info("Service startup completed")
			}
			return
		}
		if os.Args[1] == "restart" {
			if se := s.Restart(); se != nil {
				golog.Errorf("Service restart failed. %+v", se)
			} else {
				golog.Info("Service restart completed")
			}
			return
		}
		if os.Args[1] == "stop" {
			if se := s.Stop(); se != nil {
				golog.Errorf("Service stop failed. %+v", se)
			} else {
				golog.Info("Service stop completed")
			}
			return
		}
	}

	if err = s.Run(); err != nil {
		golog.Fatal(err)
	}
}

// Start 启动主服务
func Start() {
	doRun := func(server conf.Server) {
		addr := server.InetHost + ":" + strconv.Itoa(server.InetPort)
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
				AccessKey:   server.AccessKey,
				PoolHandler: pooler,
				ConnHandler: connHandler,
			}
			msgHandler.HeartBeat()

			connHandler.Listen(conn, msgHandler)
		}
	}
	// 调用
	for idx, server := range conf.AppConf.Natok.Server {
		go func(ser conf.Server) { doRun(ser) }(server)
		golog.Infof("Listen: %d, %s", idx, server.InetHost)
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
	tlsConf := conf.AppConf.Natok
	cert, err := tls.LoadX509KeyPair(tlsConf.CertPemPath, tlsConf.CertKeyPath)
	if err != nil {
		golog.Error(err)
	}
	certBytes, err := os.ReadFile(tlsConf.CertPemPath)
	if err != nil {
		golog.Fatal("Unable to read cert.pem")
	}
	clientCertPool := x509.NewCertPool()
	ok := clientCertPool.AppendCertsFromPEM(certBytes)
	if !ok {
		golog.Fatal("failed to parse root certificate")
	}
	return &tls.Config{
		RootCAs:            clientCertPool,
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
	}
}
