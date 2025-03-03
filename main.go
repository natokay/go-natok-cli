package main

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/kardianos/service"
	log "github.com/sirupsen/logrus"
	"natok-cli/conf"
	"natok-cli/core"
	"net"
	"os"
	"strconv"
	"time"
)

type Program struct{}

func (p *Program) Start(s service.Service) error {
	go p.run()
	return nil
}

func (p *Program) run() {
	log.Info("Started natok client service")
	Start()
}

func (p *Program) Stop(s service.Service) error {
	log.Info("Stop natok client service")
	return nil
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
		log.Fatal(err)
	}

	if len(os.Args) > 1 {
		if os.Args[1] == "install" {
			if se := s.Install(); se != nil {
				log.Errorf("Service installation failed. %+v", se)
			} else {
				log.Info("Service installed")
			}
			return
		}
		if os.Args[1] == "uninstall" {
			if se := s.Uninstall(); se != nil {
				log.Errorf("Service uninstall failed. %+v", se)
			} else {
				log.Info("Service uninstalled")
			}
			return
		}
		if os.Args[1] == "start" {
			if se := s.Start(); se != nil {
				log.Errorf("Service start failed. %+v", se)
			} else {
				log.Info("Service startup completed")
			}
			return
		}
		if os.Args[1] == "restart" {
			if se := s.Restart(); se != nil {
				log.Errorf("Service restart failed. %+v", se)
			} else {
				log.Info("Service restart completed")
			}
			return
		}
		if os.Args[1] == "stop" {
			if se := s.Stop(); se != nil {
				log.Errorf("Service stop failed. %+v", se)
			} else {
				log.Info("Service stop completed")
			}
			return
		}
	}

	if err = s.Run(); err != nil {
		log.Fatal(err)
	}
}

// Start 启动主服务
func Start() {
	doRun := func(server conf.Server) {
		addr := server.InetHost + ":" + strconv.Itoa(server.InetPort)
		tlsConfig := TlsConfig()
		poolHandler := &core.NatokHandler{
			Conf: &core.NatokConnConfig{
				Addr: addr,
				Conf: tlsConfig,
			},
			Conns: make([]*core.ConnectHandler, 0, 10),
		}

		connHandler := &core.ConnectHandler{Name: "Main"}

		for {
			connHandler.Conn = Connect(addr, tlsConfig)
			natokServerHandler := &core.NatokServerHandler{
				AccessKey:    server.AccessKey,
				NatokHandler: poolHandler,
				ConnHandler:  connHandler,
			}
			connHandler.MsgHandler = natokServerHandler

			natokServerHandler.HeartBeat()
			natokServerHandler.Auth()
			connHandler.Listen()
		}
	}
	// 调用
	for idx, server := range conf.AppConf.Natok.Server {
		go func(ser conf.Server) { doRun(ser) }(server)
		log.Infof("Listen: %d, %s", idx+1, server.InetHost)
	}
}

// Connect 向NATOK-SERVER发起连接
func Connect(addr string, conf *tls.Config) net.Conn {
	retry := 0
	for {
		var conn net.Conn
		var err error

		if conf != nil {
			conn, err = tls.Dial("tcp", addr, conf)
		} else {
			conn, err = net.Dial("tcp", addr)
		}

		if err != nil {
			retry += 1
			if retry > 1000000 {
				retry = 1
			}
			if retry%30 == 0 {
				log.Warnf("Connection to natok server exception! retry: %d Addr: %s, Error: %+v", retry, addr, err)
			}
			time.Sleep(time.Second * 2)
			continue
		}

		return conn
	}
}

// TlsConfig TSL协议配置
func TlsConfig() *tls.Config {
	tlsConf := conf.AppConf.Natok
	cert, err := tls.LoadX509KeyPair(tlsConf.CertPemPath, tlsConf.CertKeyPath)
	if err != nil {
		log.Error(err)
	}
	certBytes, err := os.ReadFile(tlsConf.CertPemPath)
	if err != nil {
		log.Fatal("Unable to read cert.pem")
	}
	clientCertPool := x509.NewCertPool()
	ok := clientCertPool.AppendCertsFromPEM(certBytes)
	if !ok {
		log.Fatal("failed to parse root certificate")
	}
	return &tls.Config{
		RootCAs:            clientCertPool,
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
	}
}
