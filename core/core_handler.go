package core

import (
	log "github.com/sirupsen/logrus"
	"net"
	"runtime/debug"
	"time"
)

// Message struct 消息体对象
type Message struct {
	Type   byte   // 消息类型
	Serial string // 消息序列
	Net    string // 网络类型
	Uri    string // 消息头
	Data   []byte // 消息体
}

// MsgHandler interface 消息处理接口
type MsgHandler interface {
	Error(*ConnectHandler)                //出错
	Encode(interface{}) []byte            //编码
	Decode([]byte) (interface{}, int)     //解码
	Receive(*ConnectHandler, interface{}) //接收
}

// ConnectHandler struct 通道链接载体
type ConnectHandler struct {
	Name        string          //通道名称
	ReadTime    time.Time       //读取时间
	WriteTime   time.Time       //写入时间
	Active      bool            //是否活跃
	ReadBuf     []byte          //读取的内容
	Conn        net.Conn        //连接通道
	MsgHandler  MsgHandler      //消息句柄
	ConnHandler *ConnectHandler //连接句柄
}

// Write 消息写入
func (c *ConnectHandler) Write(msg interface{}) {
	if c.MsgHandler == nil {
		return
	}
	data := c.MsgHandler.Encode(msg)
	c.WriteTime = time.Now()
	_, _ = c.Conn.Write(data)
}

// Listen 连接请求监听
func (c *ConnectHandler) Listen() {
	defer func() {
		if err := recover(); err != nil {
			c.MsgHandler.Error(c)
			log.Warnf("Warn: %+v", err)
			debug.PrintStack()
		}
	}()

	if c.Conn == nil {
		return
	}

	c.Active = true
	c.ReadTime = time.Now()

	for c.Active {
		// 最大缓冲4M
		if c.ReadBuf != nil && len(c.ReadBuf) > MaxPacketSize {
			log.Error("Warn: This conn is error ! Packet max than 4M !")
			_ = c.Conn.Close()
		}

		// 最大包64kb
		buf := make([]byte, 1024*64)
		n, err := c.Conn.Read(buf)
		if err != nil || n == 0 {
			log.Errorf("Error: %+v", err)
			c.Active = false
			c.MsgHandler.Error(c)
			return
		}

		c.ReadTime = time.Now()
		if c.ReadBuf == nil {
			c.ReadBuf = buf[0:n]
		} else {
			c.ReadBuf = append(c.ReadBuf, buf[0:n]...)
		}

		for {
			msg, n := c.MsgHandler.Decode(c.ReadBuf)
			if msg == nil {
				break
			}
			c.MsgHandler.Receive(c, msg)
			c.ReadBuf = c.ReadBuf[n:]
			if len(c.ReadBuf) == 0 {
				break
			}
		}

		if len(c.ReadBuf) > 0 {
			buf := make([]byte, len(c.ReadBuf))
			copy(buf, c.ReadBuf)
			c.ReadBuf = buf
		}
	}
}
