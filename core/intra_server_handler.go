package core

import (
	"github.com/kataras/golog"
)

// IntraServerHandler struct 内网服务处理
type IntraServerHandler struct {
	Uri          string
	AccessKey    string
	PoolHandler  *PoolHandler
	natokHandler *ConnectHandler
}

// Encode 编码消息
func (i *IntraServerHandler) Encode(msg interface{}) []byte {
	if msg == nil {
		return []byte{}
	}
	return msg.([]byte)
}

// Decode 解码消息
func (i *IntraServerHandler) Decode(buf []byte) (interface{}, int) {
	return buf, len(buf)
}

// Receive 请求接收
func (i *IntraServerHandler) Receive(connHandler *ConnectHandler, data interface{}) {
	if connHandler.ConnHandler == nil {
		return
	}
	msg := Message{Type: TypeTransfer, Data: data.([]byte)}
	connHandler.ConnHandler.Write(msg)
}

// Success 成功 交换连接通道
func (i *IntraServerHandler) Success(connHandler *ConnectHandler) {
	natokHandler, err := i.PoolHandler.Pull()
	if err != nil {
		golog.Errorf("Get Connection Uri:%s error:%+v", i.Uri, err)
		msg := Message{Type: TypeDisconnect, Uri: i.Uri}
		i.natokHandler.Write(msg)
		connHandler.Conn.Close()
	} else {
		natokHandler.ConnHandler = connHandler
		connHandler.ConnHandler = natokHandler

		msg := Message{Type: TypeConnect, Uri: i.Uri + "@" + i.AccessKey}
		natokHandler.Write(msg)
		golog.Infof("Intranet server connect success, notify natok server:%s", msg.Uri)
	}
}

// Error 错误处理
func (i *IntraServerHandler) Error(connHandler *ConnectHandler) {
	conn := connHandler.ConnHandler
	if conn != nil {
		msg := Message{Type: TypeDisconnect, Uri: i.Uri}
		conn.Write(msg)
		conn.ConnHandler = nil
	}
	connHandler.MsgHandler = nil
}

// Failure 失败处理
func (i *IntraServerHandler) Failure() {
	msg := Message{Type: TypeDisconnect, Uri: i.Uri}
	i.natokHandler.Write(msg)
}
