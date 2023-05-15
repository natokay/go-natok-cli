package core

import (
	"crypto/tls"
	"encoding/binary"
	"github.com/kataras/golog"
	"net"
	"os"
	"time"
)

// NatokServerHandler struct NATOK服务处理
type NatokServerHandler struct {
	Chan        chan struct{}
	AccessKey   string
	PoolHandler *PoolHandler
	ConnHandler *ConnectHandler
}

// NatokConfig struct 地址配置
type NatokConfig struct {
	Addr string
	Conf *tls.Config
}

// Encode 编码消息
func (n *NatokServerHandler) Encode(inMsg interface{}) []byte {
	if inMsg == nil {
		return []byte{}
	}
	msg := inMsg.(Message)
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, msg.SerialNum)

	uriBytes := []byte(msg.Uri)
	bodyLen := TypeSize + SerialNumSize + UriLenSize + len(uriBytes) + len(msg.Data)

	data := make([]byte, HeaderSize, bodyLen+HeaderSize)
	binary.BigEndian.PutUint32(data, uint32(bodyLen))

	data = append(data, msg.Type)
	data = append(data, buf...)
	data = append(data, byte(len(uriBytes)))
	data = append(data, uriBytes...)
	data = append(data, msg.Data...)
	return data
}

// Decode 解码消息
func (n *NatokServerHandler) Decode(buf []byte) (interface{}, int) {
	HeaderBytes := buf[0:HeaderSize]
	HeaderLen := binary.BigEndian.Uint32(HeaderBytes)

	if uint32(len(buf)) < HeaderLen+HeaderSize {
		return nil, 0
	}

	num := int(HeaderLen + HeaderSize)
	body := buf[HeaderSize:num]

	uriLen := uint8(body[SerialNumSize+TypeSize])
	msg := Message{
		Type:      body[0],
		SerialNum: binary.BigEndian.Uint64(body[TypeSize : SerialNumSize+TypeSize]),
		Uri:       string(body[SerialNumSize+TypeSize+UriLenSize : SerialNumSize+TypeSize+UriLenSize+uriLen]),
		Data:      body[SerialNumSize+TypeSize+UriLenSize+uriLen:],
	}
	return msg, num
}

// Receive 请求接收
func (n *NatokServerHandler) Receive(connHandler *ConnectHandler, msgData interface{}) {
	msg := msgData.(Message)
	//golog.Println("Received connect message:", msg.Uri, "=>", string(msg.Data))
	switch msg.Type {
	case TypeConnect:
		go func() {
			intraServerHandler := &IntraServerHandler{
				Uri:          msg.Uri,
				natokHandler: connHandler,
				AccessKey:    n.AccessKey,
				PoolHandler:  n.PoolHandler,
			}
			addr := string(msg.Data)
			conn, err := net.Dial("tcp", addr)
			if err != nil {
				golog.Errorf("Failed to connect intranet server! %+v", err)
				intraServerHandler.Failure()
			} else {
				connectHandler := &ConnectHandler{Conn: conn}
				connectHandler.Listen(conn, intraServerHandler)
			}
		}()
	case TypeTransfer:
		if connHandler.ConnHandler != nil {
			connHandler.ConnHandler.Write(msg.Data)
		}
	case TypeDisconnect:
		if connHandler.ConnHandler != nil {
			connHandler.ConnHandler.Conn.Close()
			connHandler.ConnHandler = nil
		}
		if n.AccessKey == "" {
			n.PoolHandler.Push(connHandler)
		}
	case typeNoAvailablePort:
		golog.Warn("There are no available ports for the natok access key.")
	case TypeDisabledAccessKey:
		golog.Info("Natok access key is disabled.")
		n.Close(connHandler)
		os.Exit(1)
	case TypeInvalidKey:
		golog.Info("Natok access key is not valid.")
		n.Close(connHandler)
		os.Exit(1)
	case TypeIsInuseKey:
		golog.Info("Natok access key is in use by other natok client.")
		golog.Info("If you want to have exclusive natok service")
		golog.Info("please visit 'www.natok.cn' for more details.")
		n.Close(connHandler)
		os.Exit(1)
	case TypeDisabledTrialClient:
		golog.Info("Your natok client is overuse.")
		golog.Info("The trial natok access key can only be used for 20 minutes in 24 hours.")
		golog.Info("If you want to have exclusive natok service")
		golog.Info("please visit 'www.natok.cn' for more details.")
		n.Close(connHandler)
		os.Exit(1)
	}
}

// Success 认证成功
func (n *NatokServerHandler) Success(connHandler *ConnectHandler) {
	if n.AccessKey == "" {
		return
	}
	msg := Message{Type: TypeAuth, Uri: n.AccessKey}
	connHandler.Write(msg)
}

// Error 错误处理
func (n *NatokServerHandler) Error(connHandler *ConnectHandler) {
	if n.Chan != nil {
		close(n.Chan)
	}
	handler := connHandler.ConnHandler
	if handler != nil {
		if handler.Conn != nil {
			handler.Conn.Close()
			handler.Conn = nil
		}
		handler.ConnHandler = nil
	}
	connHandler.ConnHandler = nil
	connHandler.MsgHandler = nil
	time.Sleep(time.Second * 3)
}

// Close 关闭连接通道
func (n *NatokServerHandler) Close(connHandler *ConnectHandler) {
	if n.Chan != nil {
		close(n.Chan)
	}
	if connHandler.ConnHandler != nil {
		if connHandler.ConnHandler.Conn != nil {
			connHandler.ConnHandler.Conn.Close()
			connHandler.ConnHandler.Conn = nil
			connHandler.ConnHandler.ConnHandler = nil
		}
		connHandler.ConnHandler = nil
		//handler.Conn.Close()
	}
	if n.ConnHandler != nil && n.ConnHandler.Conn != nil {
		n.ConnHandler.Conn.Close()
		n.ConnHandler.Conn = nil
	}
	n.ConnHandler = nil
	connHandler.MsgHandler = nil
}

// HeartBeat 发送心跳包 -> NATOK-SERVER
func (n *NatokServerHandler) HeartBeat() {
	n.Chan = make(chan struct{})
	go func() {
		for {
			select {
			case <-time.After(time.Second * HeartbeatInterval):
				if time.Now().Unix()-n.ConnHandler.ReadTime >= 2*TypeHeartbeat {
					golog.Error("Natok connection timeout")
					if n.ConnHandler != nil && n.ConnHandler.Conn != nil {
						n.ConnHandler.Conn.Close()
					}
					return
				}
				msg := Message{Type: TypeHeartbeat}
				n.ConnHandler.Write(msg)
			case <-n.Chan:
				return
			}
		}
	}()
}
