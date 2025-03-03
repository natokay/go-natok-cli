package core

import (
	"crypto/tls"
	"encoding/binary"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net"
	"os"
	"time"
)

// NatokServerHandler struct NATOK服务处理
type NatokServerHandler struct {
	Chan         chan struct{}
	AccessKey    string //密钥
	source       string //来源
	target       string //目标
	NatokHandler *NatokHandler
	ConnHandler  *ConnectHandler
}

// NatokConfig struct 地址配置
type NatokConfig struct {
	Addr string
	Conf *tls.Config
}

// NatokHandler struct // Natok句柄
type NatokHandler struct {
	count Counter           //数量
	Conf  *NatokConnConfig  //配置
	Conns []*ConnectHandler //连接
}

type NatokConnConfig struct {
	Addr string
	Conf *tls.Config
}

// Get 获取连接
func (p *NatokConnConfig) Get() (*ConnectHandler, error) {
	var conn net.Conn
	var err error
	retries := 5
	dialer := net.Dialer{Timeout: 2 * time.Second}
	for tries := 0; tries <= retries; tries++ {
		if p.Conf != nil {
			conn, err = tls.DialWithDialer(&dialer, "tcp", p.Addr, p.Conf)
		} else {
			conn, err = dialer.Dial("tcp", p.Addr)
		}
		if err == nil {
			break
		}
		if tries > 0 {
			log.Warnf("Connect natok-server failed, Retries: %d/%d, Error: %+v", tries, retries, err)
		}
		time.Sleep(200 * time.Millisecond)
	}
	// 如果未连接成功
	if err != nil {
		log.Errorf("Connect natok-server failed, Error: %+v", err)
		return nil, err
	}
	connHandler := &ConnectHandler{
		Name:   "natok-server-子集",
		Active: true,
		Conn:   conn,
	}
	return connHandler, nil
}

// Encode 编码消息
func (s *NatokServerHandler) Encode(inMsg interface{}) []byte {
	if inMsg == nil {
		return []byte{}
	}
	msg := inMsg.(Message)
	serialBytes := []byte(msg.Serial)
	netBytes := []byte(msg.Net)
	UriBytes := []byte(msg.Uri)
	// byte=Uint8Size,3个string=Uint8Size*3,+data
	dataLen := Uint8Size + Uint8Size*3 + len(serialBytes) + len(netBytes) + len(UriBytes) + len(msg.Data)
	data := make([]byte, Uint32Size, Uint32Size+dataLen)
	binary.BigEndian.PutUint32(data, uint32(dataLen))

	data = append(data, msg.Type)
	data = append(data, byte(len(serialBytes)))
	data = append(data, byte(len(netBytes)))
	data = append(data, byte(len(UriBytes)))
	data = append(data, serialBytes...)
	data = append(data, netBytes...)
	data = append(data, UriBytes...)
	data = append(data, msg.Data...)
	return data
}

// Decode 解码消息
func (s *NatokServerHandler) Decode(buf []byte) (interface{}, int) {
	headerBytes := buf[0:Uint32Size]
	headerLen := binary.BigEndian.Uint32(headerBytes)
	// 来自客户端的包，校验完整性。
	if uint32(len(buf)) < headerLen+Uint32Size {
		return nil, 0
	}

	head := int(Uint32Size + headerLen)
	body := buf[Uint32Size:head]
	serialLen := int(body[Uint8Size])
	netLen := int(body[Uint8Size*2])
	uriLen := int(body[Uint8Size*3])
	msg := Message{
		Type:   body[0],
		Serial: string(body[Uint8Size*4 : Uint8Size*4+serialLen]),
		Net:    string(body[Uint8Size*4+serialLen : Uint8Size*4+serialLen+netLen]),
		Uri:    string(body[Uint8Size*4+serialLen+netLen : Uint8Size*4+serialLen+netLen+uriLen]),
		Data:   body[Uint8Size*4+serialLen+netLen+uriLen:],
	}
	return msg, head
}

// Receive 请求接收
func (s *NatokServerHandler) Receive(connHandler *ConnectHandler, msgData interface{}) {
	msg := msgData.(Message)
	//log.Println("Received connect message:", msg.Uri, "=>", string(msg.Data))
	switch msg.Type {
	// 连接到natok服务
	case TypeConnectNatok:
		go func() {
			log.Debugf("1-1 ===== From natok server message: %s %s", msg.Serial, string(msg.Data))
			if natokHandler, err := s.NatokHandler.Conf.Get(); err == nil {
				natokServerHandler := &NatokServerHandler{
					AccessKey:   s.AccessKey,
					ConnHandler: natokHandler,
				}
				log.Debugf("1-2 =====Connect natok, Listen natok server message: %s %s", msg.Serial, string(msg.Data))
				natokHandler.MsgHandler = natokServerHandler
				natokServerHandler.HeartBeat()
				natokHandler.Write(Message{Type: TypeConnectNatok, Serial: msg.Serial, Uri: s.AccessKey})
				natokHandler.Listen()
				log.Debugf("1-3 =====Disconnect natok, Listen natok server message: %s %s", msg.Serial, string(msg.Data))
			} else {
				log.Errorf("1-e =====Connect natok server failed, Message: %s %s, Error: %+v", msg.Serial, string(msg.Data), err)
			}
		}()
	// 连接到内部服务
	case TypeConnectIntra:
		go func() {
			network := msg.Net
			addr := string(msg.Data)
			s.source = fmt.Sprintf("%s://%s", network, addr)
			s.target = msg.Uri
			sprintf := fmt.Sprintf("%s %s -> %s", msg.Serial, s.source, s.target)
			log.Debugf("2-1 ===== From natok server message: %s", sprintf)
			if conn, err := net.Dial(network, addr); err == nil {
				intraHandler := &ConnectHandler{Name: network + addr, Conn: conn, Active: true, ConnHandler: connHandler}
				intraHandler.MsgHandler = &IntraServerHandler{
					Uri:            msg.Uri,
					AccessKey:      s.AccessKey,
					connectHandler: connHandler,
				}
				connHandler.ConnHandler = intraHandler
				connHandler.Write(Message{Type: TypeConnectIntra, Serial: msg.Serial, Uri: s.AccessKey})
				log.Debugf("2-2 =====Connect intranet, Listen natok server message: %s", sprintf)
				intraHandler.Listen()
				log.Debugf("2-3 =====Disconnect intranet, Listen natok server message: %s", sprintf)
			} else {
				log.Errorf("2-e =====Connect intranet server failed, Message: %s, Error: %+v", sprintf, err)
			}
		}()
	// 传输数据 - 转发内部服务
	case TypeTransfer:
		sprintf := fmt.Sprintf("%s %s -> %s", msg.Serial, s.source, s.target)
		log.Debugf("3-1 =====TypeTransfer natok server message: %s", sprintf)
		if conn := connHandler.ConnHandler; conn != nil {
			log.Debugf("3-2 =====TypeTransfer intranet server message: %s", sprintf)
			conn.Write(msg.Data)
		}
	// 关闭连接 - 断开内部服务
	case TypeDisconnect:
		sprintf := fmt.Sprintf("%s %s -> %s", msg.Serial, s.source, s.target)
		log.Debugf("4-1 =====TypeDisconnect natok server message: %s", sprintf)
		if conn := connHandler.ConnHandler; conn != nil {
			_ = conn.Conn.Close()
			connHandler.ConnHandler = nil
		}
	case typeNoAvailablePort:
		log.Warnf("Natok access key %s no available ports.", msg.Uri)
	case TypeDisabledAccessKey:
		log.Warnf("Natok access key %s is disabled.", msg.Uri)
	case TypeInvalidKey:
		log.Errorf("Natok access key %s is not valid.", msg.Uri)
		s.Close(connHandler)
		os.Exit(1)
	case TypeIsInuseKey:
		log.Warnf("Natok access key %s is in use by other natok client.", msg.Uri)
		s.Close(connHandler)
		os.Exit(1)
	case TypeDisabledTrialClient:
		log.Infof("Natok access key %s is overuse.", msg.Uri)
		s.Close(connHandler)
		os.Exit(1)
	}
}

// Auth 认证成功
func (s *NatokServerHandler) Auth() {
	if s.AccessKey == "" {
		return
	}
	msg := Message{Type: TypeAuth, Serial: "1", Net: "tcp", Uri: s.AccessKey, Data: []byte("8888")}
	s.ConnHandler.Write(msg)
}

// Error 错误处理
func (s *NatokServerHandler) Error(connHandler *ConnectHandler) {
	if s.Chan != nil {
		close(s.Chan)
	}
	intraHandler := connHandler.ConnHandler
	if intraHandler != nil {
		if conn := intraHandler.Conn; conn != nil {
			_ = conn.Close()
		}
		intraHandler.ConnHandler = nil
	}
	connHandler.ConnHandler = nil
	connHandler.MsgHandler = nil
	time.Sleep(time.Second * 3)
}

// Close 关闭连接通道
func (s *NatokServerHandler) Close(connHandler *ConnectHandler) {
	if s.Chan != nil {
		close(s.Chan)
	}
	if intraHandler := connHandler.ConnHandler; intraHandler != nil {
		if intraHandler.Conn != nil {
			intraHandler.Active = false
			_ = intraHandler.Conn.Close()
			intraHandler.Conn = nil
			intraHandler.ConnHandler = nil
		}
		connHandler.ConnHandler = nil
	}
	if connHandler.Conn != nil {
		connHandler.Active = false
		_ = connHandler.Conn.Close()
		connHandler.Conn = nil
	}
	s.ConnHandler = nil
	connHandler.MsgHandler = nil
}

// HeartBeat 发送心跳包 -> NATOK-SERVER
func (s *NatokServerHandler) HeartBeat() {
	s.Chan = make(chan struct{})
	go func() {
		for {
			now := time.Now()
			select {
			case <-time.After(10 * time.Second):
				// 若通道在30s内未收到过数据，则发送一次心跳包。
				if now.Sub(s.ConnHandler.ReadTime) >= 30*time.Second {
					msg := Message{Type: TypeHeartbeat, Uri: s.AccessKey}
					s.ConnHandler.Write(msg)
				}
			case <-s.Chan:
				return
			}
		}
	}()
}
