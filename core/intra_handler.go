package core

import log "github.com/sirupsen/logrus"

// IntraServerHandler struct 内网服务处理
type IntraServerHandler struct {
	Uri            string
	AccessKey      string
	NatokHandler   *NatokHandler
	connectHandler *ConnectHandler
}

// Encode 编码消息
func (s *IntraServerHandler) Encode(msg interface{}) []byte {
	if msg == nil {
		return []byte{}
	}
	return msg.([]byte)
}

// Decode 解码消息
func (s *IntraServerHandler) Decode(buf []byte) (interface{}, int) {
	return buf, len(buf)
}

// Receive 请求接收
func (s *IntraServerHandler) Receive(connHandler *ConnectHandler, data interface{}) {
	if conn := connHandler.ConnHandler; conn != nil {
		msg := Message{Type: TypeTransfer, Data: data.([]byte)}
		conn.Write(msg)
		log.Debugf("intra receive message %s", connHandler.Name)
	}
}

// Error 错误处理
func (s *IntraServerHandler) Error(connHandler *ConnectHandler) {
	if natokHandler := connHandler.ConnHandler; natokHandler != nil {
		msg := Message{Type: TypeDisconnect, Uri: s.Uri}
		natokHandler.Write(msg)
		connHandler.ConnHandler = nil
	}
}

// Failure 失败处理
func (s *IntraServerHandler) Failure() {
	msg := Message{Type: TypeDisconnect, Uri: s.Uri}
	s.connectHandler.Write(msg)
}
